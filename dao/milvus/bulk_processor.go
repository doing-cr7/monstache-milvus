package milvus

import (
	"context"
	"errors"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cenkalti/backoff/v5"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// var infoLog = log.New(os.Stdout, "INFO ", log.Flags())
// var warnLog = log.New(os.Stdout, "WARN ", log.Flags())
// var statsLog = log.New(os.Stdout, "STATS ", log.Flags())
// var traceLog = log.New(os.Stdout, "TRACE ", log.Flags())
var errorLog = log.New(os.Stderr, "ERROR ", log.Flags())

// --- 类型定义 ---

type MilvusBeforeFunc func(executionId int64, requests []any)
type MilvusAfterFunc func(executionId int64, requests []any, err error)

// MilvusBulkProcessorService 模仿 client.BulkProcessor() 的 Builder
type MilvusBulkProcessorService struct {
	client         *milvusclient.Client
	collectionName string
	numWorkers     int
	bulkActions    int
	retry          bool
	bulkSize       int
	flushInterval  time.Duration
	wantStats      bool
	beforeFn       MilvusBeforeFunc
	afterFn        MilvusAfterFunc
}

// MilvusBulkProcessor 真正的执行器
type MilvusBulkProcessor struct {
	client         *milvusclient.Client
	collectionName string
	numWorkers     int
	bulkActions    int
	bulkSize       int
	flushInterval  time.Duration
	beforeFn       MilvusBeforeFunc
	afterFn        MilvusAfterFunc
	retry          bool
	wantStats      bool

	executionId int64
	requestsC   chan any
	workerWg    sync.WaitGroup
	closeOnce   sync.Once   // 保护 Close() 方法，防止重复调用
	closed      atomic.Bool // 标记是否已关闭
}

// --- Service 构建器方法 ---

func NewMilvusBulkProcessorService(c *milvusclient.Client, collection string) *MilvusBulkProcessorService {
	return &MilvusBulkProcessorService{
		client:         c,
		collectionName: collection,
		numWorkers:     1,
		bulkActions:    1000,
		bulkSize:       5 << 20, // 5MB
		flushInterval:  1 * time.Second,
		retry:          true,
	}
}

func (s *MilvusBulkProcessorService) Name(name string) *MilvusBulkProcessorService {
	s.collectionName = name
	return s
}

func (s *MilvusBulkProcessorService) Workers(n int) *MilvusBulkProcessorService {
	s.numWorkers = n
	return s
}

func (s *MilvusBulkProcessorService) BulkActions(n int) *MilvusBulkProcessorService {
	s.bulkActions = n
	return s
}

func (s *MilvusBulkProcessorService) BulkSize(n int) *MilvusBulkProcessorService {
	s.bulkSize = n
	return s
}

func (s *MilvusBulkProcessorService) FlushInterval(d time.Duration) *MilvusBulkProcessorService {
	s.flushInterval = d
	return s
}

func (s *MilvusBulkProcessorService) Stats(want bool) *MilvusBulkProcessorService {
	s.wantStats = want
	return s
}

func (s *MilvusBulkProcessorService) DisableRetry(disable bool) *MilvusBulkProcessorService {
	s.retry = !disable
	return s
}

func (s *MilvusBulkProcessorService) Before(fn MilvusBeforeFunc) *MilvusBulkProcessorService {
	s.beforeFn = fn
	return s
}

func (s *MilvusBulkProcessorService) After(fn MilvusAfterFunc) *MilvusBulkProcessorService {
	s.afterFn = fn
	return s
}

func (s *MilvusBulkProcessorService) Do(ctx context.Context) (*MilvusBulkProcessor, error) {
	p := &MilvusBulkProcessor{
		client:         s.client,
		collectionName: s.collectionName,
		numWorkers:     s.numWorkers,
		bulkActions:    s.bulkActions,
		bulkSize:       s.bulkSize,
		flushInterval:  s.flushInterval,
		retry:          s.retry,
		wantStats:      s.wantStats,
		beforeFn:       s.beforeFn,
		afterFn:        s.afterFn,
		requestsC:      make(chan any, 2048),
	}
	p.start(ctx)
	return p, nil
}

// --- Processor 执行方法 ---

func (p *MilvusBulkProcessor) start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		p.workerWg.Add(1)
		go p.workerLoop(ctx, i)
	}
}

func (p *MilvusBulkProcessor) Add(request any) error {
	// 检查是否已关闭，避免向已关闭的 channel 发送数据导致 panic
	if p.closed.Load() {
		return errors.New("processor is closed")
	}

	// 使用 defer recover 作为最后一道防线，捕获向已关闭 channel 发送数据的 panic
	defer func() {
		if r := recover(); r != nil {
			errorLog.Printf("MilvusBulkProcessor: panic in Add() - %v", r)
		}
	}()

	// 直接发送到 channel，如果 channel 满了会阻塞等待
	// 这是正常的流控行为，让生产者等待消费者处理
	p.requestsC <- request
	return nil
}

func (p *MilvusBulkProcessor) Close() error {
	// 使用 sync.Once 确保只关闭一次
	p.closeOnce.Do(func() {
		p.closed.Store(true)
		close(p.requestsC)
	})
	p.workerWg.Wait()
	return nil
}

func (p *MilvusBulkProcessor) workerLoop(ctx context.Context, workerId int) {
	defer p.workerWg.Done()

	var (
		batch        []any
		currentBytes int
		ticker       = time.NewTicker(p.flushInterval)
	)
	defer ticker.Stop()

	commit := func() {
		if len(batch) == 0 {
			return
		}

		id := atomic.AddInt64(&p.executionId, 1)
		reqs := batch

		if p.beforeFn != nil {
			p.beforeFn(id, reqs)
		}

		var err error
		op := func() (struct{}, error) {
			_, err := p.client.Upsert(ctx, milvusclient.NewRowBasedInsertOption(p.collectionName, reqs...))
			return struct{}{}, err
		}
		if p.retry {
			notify := func(err error, d time.Duration) {
				errorLog.Printf("[Worker %d] Batch %d failed, retrying in %v. Error: %v\n",
					workerId, id, d, err)
			}
			_, err = backoff.Retry(ctx, op,
				backoff.WithBackOff(backoff.NewExponentialBackOff()),
				backoff.WithMaxTries(5),
				backoff.WithNotify(notify),
			)
		} else {
			_, err = op()
		}

		if p.afterFn != nil {
			p.afterFn(id, reqs, err)
		}

		if err != nil {
			errorLog.Printf("[Worker %d] Batch %d FAILED permanently after retries. Error: %v\n", workerId, id, err)
		}

		batch = make([]any, 0, p.bulkActions)
		currentBytes = 0
	}

	for {
		select {
		case req, open := <-p.requestsC:
			if !open {
				commit()
				return
			}
			batch = append(batch, req)

			currentBytes += estimateRequestSize(req)
			if (p.bulkActions > 0 && len(batch) >= p.bulkActions) ||
				(p.bulkSize > 0 && currentBytes >= p.bulkSize) {
				commit()
			}
		case <-ticker.C:
			commit()
		case <-ctx.Done():

			commit()
			return
		}
	}
}

func estimateRequestSize(req any) int {
	if req == nil {
		return 0
	}

	v := reflect.ValueOf(req)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}

		return int(unsafe.Sizeof(req)) + estimateRequestSize(v.Elem().Interface())

	case reflect.Map:

		baseSize := int(unsafe.Sizeof(req))
		if v.Len() == 0 {
			return baseSize
		}

		sampleSize := 5
		if v.Len() < sampleSize {
			sampleSize = v.Len()
		}

		totalSampleSize := 0
		count := 0
		for _, key := range v.MapKeys() {
			if count >= sampleSize {
				break
			}
			totalSampleSize += estimateRequestSize(key.Interface())
			totalSampleSize += estimateRequestSize(v.MapIndex(key).Interface())
			count++
		}

		if count > 0 {
			avgPairSize := totalSampleSize / count
			return baseSize + avgPairSize*v.Len()
		}
		return baseSize + v.Len()*64

	case reflect.Slice, reflect.Array:
		size := int(unsafe.Sizeof(req))
		if v.Len() > 0 {

			elemSize := int(v.Type().Elem().Size())
			size += v.Len() * elemSize
		}
		return size

	case reflect.String:
		return int(unsafe.Sizeof(req)) + v.Len()

	case reflect.Struct:

		return int(v.Type().Size())

	default:

		return int(unsafe.Sizeof(req))
	}
}
