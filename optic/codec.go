package optic

// Decoder is an interface defining functions that a decoder plugin must satisfy.
type Decoder interface {
	SetEventType(eventType EventType) error

	Decode(src []byte) ([]Event, error)

	DecodeLine(line string) (Event, error)
}

// DecoderInput is an interface for source plugins that are able to decode
// arbitrary data formats.
type DecoderInput interface {
	// SetDecoder sets the decoder function for the interface.
	SetDecoder(decoder Decoder)
}

// Encoder is an interface defining functions that a encoder plugin must satisfy.
type Encoder interface {
	Encode(event Event) ([]byte, error)

	EncodeTo(event Event, dst []byte) error
}

// EncoderOutput is an interface for sink plugins that are able to encode optic
// events into arbitrary data formats.
type EncoderOutput interface {
	// SetEncoder sets the encoder function for the interface.
	SetEncoder(encoder Encoder)
}

// Codec is an interface that joins `Decoder` and `Encoder` together.
type Codec interface {
	Decoder
	Encoder
}
