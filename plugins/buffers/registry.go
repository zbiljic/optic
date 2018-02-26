package buffers

import (
	"fmt"

	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/zbiljic/optic/optic"
)

type creator func() optic.Buffer

var buffers = map[string]creator{}

func Add(name string, creator creator) {
	buffers[name] = creator
}

func NewDefaultBuffer() (optic.Buffer, error) {
	return NewBuffer(map[string]interface{}{"kind": "memory"})
}

func NewBuffer(config map[string]interface{}) (optic.Buffer, error) {
	var (
		buffer optic.Buffer
		err    error
	)

	var kind string
	kind, err = cast.ToStringE(config["kind"])
	if err != nil || kind == "" {
		err = fmt.Errorf("Undefined buffer kind")
		return nil, err
	}

	var (
		creator creator
		ok      bool
	)
	if creator, ok = buffers[kind]; !ok {
		err = fmt.Errorf("Invalid buffer kind: %s", kind)
		return nil, err
	}
	buffer = creator()

	delete(config, "kind")

	// unmarshal configuration for concrete buffer
	var lv = viper.New()
	lv.Set("config", config)
	err = lv.UnmarshalKey("config", buffer)
	if err != nil {
		return nil, err
	}

	// build the buffer
	err = buffer.Build()

	return buffer, err
}
