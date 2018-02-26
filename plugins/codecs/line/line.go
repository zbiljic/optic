package line

import (
	"bytes"
	"fmt"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/optic/raw"
	"github.com/zbiljic/optic/plugins/codecs"
)

const (
	name = "line"
)

type LineCodec struct {
	// EventType is the type of the event that will be produced by this codec.
	EventType optic.EventType `mapstructure:"-"`

	// DefaultTags will be added to every decoded event.
	DefaultTags map[string]string `mapstructure:"tags"`
}

func NewLineCodec() optic.Codec {
	return &LineCodec{
		EventType:   optic.RawEvent,
		DefaultTags: make(map[string]string),
	}
}

func (c *LineCodec) SetEventType(eventType optic.EventType) error {
	switch eventType {
	case optic.RawEvent:
		c.EventType = eventType
	default:
		return fmt.Errorf("%s codec does not support %s event type",
			name, eventType)
	}
	return nil
}

func (c *LineCodec) Decode(src []byte) ([]optic.Event, error) {
	events := make([]optic.Event, 0)

	switch c.EventType {
	case optic.RawEvent:
		lines := bytes.Split(src, []byte{'\n'})
		for _, line := range lines {
			raw, err := raw.New(name, line, c.DefaultTags, nil)
			if err != nil {
				return nil, err
			}
			events = append(events, raw)
		}
	default:
		return nil, fmt.Errorf("%s codec does not support %s event type",
			name, c.EventType)
	}

	return events, nil
}

func (c *LineCodec) DecodeLine(line string) (optic.Event, error) {

	events, err := c.Decode([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(events) < 1 {
		return nil, fmt.Errorf("Can not decode line: [%s], for codec: %s", line, name)
	}

	return events[0], nil
}

func (c *LineCodec) Encode(event optic.Event) ([]byte, error) {
	out := []byte{}
	var err error

	switch v := event.(type) {
	case optic.Raw:
		buf := v.Serialize()
		out = append(out, buf...)
		out = append(out, '\n')
	case optic.Metric:
		buf := v.Serialize()
		out = append(out, buf...)
		out = append(out, '\n')
	case optic.LogLine:
		buf := v.Serialize()
		out = append(out, buf...)
		out = append(out, '\n')
	default:
		err = fmt.Errorf("%s codec does not support %s event type",
			name, v.Type())
	}

	return out, err
}

func (c *LineCodec) EncodeTo(event optic.Event, dst []byte) error {
	var err error

	switch v := event.(type) {
	case optic.Raw:
		buf := v.Serialize()
		buf = append(buf, '\n')
		copy(dst, buf)
	case optic.Metric:
		buf := v.Serialize()
		buf = append(buf, '\n')
		copy(dst, buf)
	case optic.LogLine:
		buf := v.Serialize()
		buf = append(buf, '\n')
		copy(dst, buf)
	default:
		err = fmt.Errorf("%s codec does not support %s event type",
			name, c.EventType)
	}

	return err
}

func init() {
	codecs.Add(name, NewLineCodec)
}
