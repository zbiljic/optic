package codecs

import (
	"fmt"

	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/zbiljic/optic/optic"
)

type creator func() optic.Codec

var codecs = map[string]creator{}

func Add(name string, creator creator) {
	codecs[name] = creator
}

func NewCodec(config map[string]interface{}) (optic.Codec, error) {
	var (
		codec optic.Codec
		err   error
	)

	var kind string
	kind, err = cast.ToStringE(config["kind"])
	if err != nil || kind == "" {
		err = fmt.Errorf("Undefined codec kind")
		return nil, err
	}

	var (
		creator creator
		ok      bool
	)
	if creator, ok = codecs[kind]; !ok {
		err = fmt.Errorf("Invalid codec kind: %s", kind)
		return nil, err
	}
	codec = creator()

	var eventTypeString string
	eventTypeString, err = cast.ToStringE(config["event"])
	if err != nil {
		err = fmt.Errorf("Invalid codec event data type for: %s", kind)
		return nil, err
	}
	var eventType optic.EventType
	switch eventTypeString {
	case "":
		break
	case "raw":
		eventType = optic.RawEvent
	case "metric":
		eventType = optic.MetricEvent
	case "logline":
		eventType = optic.LogLineEvent
	default:
		return nil, fmt.Errorf("Invalid event type: '%s'", eventTypeString)
	}
	if eventTypeString != "" {
		err = codec.SetEventType(eventType)
		if err != nil {
			return nil, err
		}
	}

	delete(config, "kind")
	delete(config, "event")

	// unmarshal configuration for concrete codec
	var lv = viper.New()
	lv.Set("config", config)
	err = lv.UnmarshalKey("config", codec)

	return codec, err
}
