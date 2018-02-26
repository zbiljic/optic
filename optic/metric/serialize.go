package metric

import (
	"fmt"
	"strconv"

	"github.com/zbiljic/optic/optic"
)

const maxInt = int(^uint(0) >> 1)

func serialize(metric optic.Metric) []byte {
	b := []byte(metric.Name())
	for k, v := range metric.Tags() {
		b = append(b, ',')
		b = appendTag(b, k, v)
	}
	b = append(b, ' ')
	for k, v := range metric.Fields() {
		b = appendField(b, k, v)
		b = append(b, ' ')
	}
	b = append(b, []byte(fmt.Sprint(metric.Time().UnixNano()))...)
	return b
}

func appendTag(b []byte, k, v string) []byte {
	b = append(b, []byte(k)...)
	b = append(b, '=')
	b = append(b, []byte(v)...)
	return b
}

func appendField(b []byte, k string, v interface{}) []byte {
	if v == nil {
		return b
	}
	b = append(b, []byte(k)...)
	b = append(b, '=')

	// check popular types first
	switch v := v.(type) {
	case float64:
		b = strconv.AppendFloat(b, v, 'f', -1, 64)
	case int64:
		b = strconv.AppendInt(b, v, 10)
		b = append(b, 'i')
	case string:
		b = append(b, '"')
		b = append(b, []byte(sanitize(v, "fieldval"))...)
		b = append(b, '"')
	case bool:
		b = strconv.AppendBool(b, v)
	case int32:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int16:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int8:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint64:
		// Cap uints above the maximum int value
		var intv int64
		if v <= uint64(maxInt) {
			intv = int64(v)
		} else {
			intv = int64(maxInt)
		}
		b = strconv.AppendInt(b, intv, 10)
		b = append(b, 'i')
	case uint32:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint16:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint8:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint:
		// Cap uints above the maximum int value
		var intv int64
		if v <= uint(maxInt) {
			intv = int64(v)
		} else {
			intv = int64(maxInt)
		}
		b = strconv.AppendInt(b, intv, 10)
		b = append(b, 'i')
	case float32:
		b = strconv.AppendFloat(b, float64(v), 'f', -1, 32)
	case []byte:
		b = append(b, v...)
	default:
		// Can't determine the type, so convert to string
		b = append(b, '"')
		b = append(b, []byte(sanitize(fmt.Sprintf("%v", v), "fieldval"))...)
		b = append(b, '"')
	}

	return b
}
