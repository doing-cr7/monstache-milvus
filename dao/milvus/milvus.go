package milvus

import (
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type ZillizMilvusClient struct {
	*milvusclient.Client
}

func NewClient(c *milvusclient.Client) *ZillizMilvusClient {
	return &ZillizMilvusClient{Client: c}
}

func (c *ZillizMilvusClient) BulkProcessor(collection string) *MilvusBulkProcessorService {
	return NewMilvusBulkProcessorService(c.Client, collection)
}
