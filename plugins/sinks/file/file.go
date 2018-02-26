package file

import (
	"fmt"
	"io"
	"os"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/codecs/line"
	"github.com/zbiljic/optic/plugins/sinks"
)

const (
	name        = "file"
	description = `Send events to file(s).`
)

type File struct {
	Files []string `mapstructure:"files"`

	encoder optic.Encoder

	writer  io.Writer
	closers []io.Closer
}

func NewFile() optic.Sink {
	codec := line.NewLineCodec()
	return &File{
		encoder: codec,
	}
}

func (*File) Kind() string {
	return name
}

func (*File) Description() string {
	return description
}

func (f *File) Connect() error {
	writers := []io.Writer{}

	if len(f.Files) == 0 {
		f.Files = []string{"stdout"}
	}

	for _, file := range f.Files {
		switch file {
		case "stdout":
			writers = append(writers, os.Stdout)
		case "stderr":
			writers = append(writers, os.Stderr)
		default:
			var of *os.File
			var err error
			if _, err := os.Stat(file); os.IsNotExist(err) {
				of, err = os.Create(file)
			} else {
				of, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			}

			if err != nil {
				return err
			}
			writers = append(writers, of)
			f.closers = append(f.closers, of)
		}
	}

	f.writer = io.MultiWriter(writers...)
	return nil
}

func (f *File) Close() error {
	var errS string
	for _, c := range f.closers {
		if err := c.Close(); err != nil {
			errS += err.Error() + "\n"
		}
	}
	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

func (f *File) Write(events []optic.Event) error {
	for _, event := range events {
		b, err := f.encoder.Encode(event)
		if err != nil {
			return fmt.Errorf("failed to encode event: %s", err)
		}
		_, err = f.writer.Write(b)
		if err != nil {
			return fmt.Errorf("failed to write event: %s, %s", event.String(), err)
		}
	}
	return nil
}

func (f *File) SetEncoder(encoder optic.Encoder) {
	f.encoder = encoder
}

func init() {
	sinks.Add(name, NewFile)
}
