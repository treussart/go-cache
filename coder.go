package cache

import (
	"fmt"

	json "github.com/goccy/go-json"
	"github.com/vmihailenco/msgpack/v5"
)

// Coder defines the interface for encoding and decoding cache values.
type Coder interface {
	Encode(value any) ([]byte, error)
	Decode(data []byte, value any) error
}

// MsgPackCoder encodes and decodes cache values using MessagePack serialization.
type MsgPackCoder struct{}

// Decode deserializes MessagePack data into value.
func (*MsgPackCoder) Decode(data []byte, value any) error {
	err := msgpack.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("msgpack.Unmarshal: %w", err)
	}
	return nil
}

// Encode serializes value into MessagePack bytes.
func (*MsgPackCoder) Encode(value any) ([]byte, error) {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("msgpack.Marshal: %w", err)
	}
	return b, nil
}

// JSONCoder encodes and decodes cache values using JSON serialization.
type JSONCoder struct{}

// Decode deserializes JSON data into value.
func (*JSONCoder) Decode(data []byte, value any) error {
	err := json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}
	return nil
}

// Encode serializes value into JSON bytes.
func (*JSONCoder) Encode(value any) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}
	return b, nil
}
