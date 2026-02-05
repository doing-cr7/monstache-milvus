package milvus

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const MilvusTitleVectorDim = 256

var zillizStringFields = []string{
	"Url",
	"Title",
}

func BuildZillizFixedRow(data map[string]interface{}, idFallback string) (map[string]interface{}, error) {
	if idFallback == "" {
		return nil, errors.New("zilliz field id missing value")
	}
	idString := idFallback

	vectorValue, ok := data["title_vector"]
	if !ok || vectorValue == nil {
		return nil, errors.New("zilliz field title_vector missing value")
	}
	vector, ok := toFloat32Slice(vectorValue)
	if !ok {
		return nil, errors.New("zilliz field title_vector expects float vector")
	}
	if len(vector) != MilvusTitleVectorDim {
		return nil, fmt.Errorf("zilliz field title_vector expects dim %d", MilvusTitleVectorDim)
	}

	row := map[string]interface{}{
		"id":           idString,
		"title_vector": vector,
	}

	if raw, ok := data["is_active"]; ok && raw != nil {
		val, ok := toInt64Value(raw)
		if !ok {
			return nil, errors.New("zilliz field is_active expects int64")
		}
		row["is_active"] = val
	}

	for _, name := range zillizStringFields {
		raw, ok := data[name]
		if !ok || raw == nil {
			row[name] = ""
			continue
		}
		val, err := zillizStringValue(raw)
		if err != nil {
			return nil, err
		}
		row[name] = val
	}

	return row, nil
}

func toInt64Value(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return parsed, true
		}
	case primitive.Decimal128:
		parsed, err := strconv.ParseInt(v.String(), 10, 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func toFloat32Slice(value interface{}) ([]float32, bool) {
	switch v := value.(type) {
	case []float32:
		return v, true
	case []float64:
		out := make([]float32, len(v))
		for i, val := range v {
			out[i] = float32(val)
		}
		return out, true
	case []interface{}:
		out := make([]float32, len(v))
		for i, val := range v {
			fv, ok := toFloat32(val)
			if !ok {
				return nil, false
			}
			out[i] = fv
		}
		return out, true
	}
	return nil, false
}

func toFloat32(value interface{}) (float32, bool) {
	switch v := value.(type) {
	case float32:
		return v, true
	case float64:
		return float32(v), true
	case int:
		return float32(v), true
	case int32:
		return float32(v), true
	case int64:
		return float32(v), true
	case string:
		parsed, err := strconv.ParseFloat(v, 32)
		if err == nil {
			return float32(parsed), true
		}
	}
	return 0, false
}

func zillizStringValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case primitive.ObjectID:
		return v.Hex(), nil
	default:
		raw, err := json.Marshal(v)
		if err == nil {
			return string(raw), nil
		}
		return fmt.Sprintf("%v", v), nil
	}
}
