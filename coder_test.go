package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Name  string `json:"name" msgpack:"name"`
	Age   int    `json:"age" msgpack:"age"`
	Email string `json:"email" msgpack:"email"`
}

func TestMsgPackCoder_roundtrip(t *testing.T) {
	coder := &MsgPackCoder{}
	original := testPayload{Name: "Alice", Age: 30, Email: "alice@example.com"}

	data, err := coder.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded testPayload
	err = coder.Decode(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestMsgPackCoder_roundtrip_slice(t *testing.T) {
	coder := &MsgPackCoder{}
	original := []string{"a", "b", "c"}

	data, err := coder.Encode(original)
	require.NoError(t, err)

	var decoded []string
	err = coder.Decode(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestMsgPackCoder_Decode_invalid(t *testing.T) {
	coder := &MsgPackCoder{}
	var dest testPayload
	err := coder.Decode([]byte("not valid msgpack{{{"), &dest)
	require.Error(t, err)
	require.ErrorContains(t, err, "msgpack.Unmarshal")
}

func TestMsgPackCoder_Encode_unsupported(t *testing.T) {
	coder := &MsgPackCoder{}
	_, err := coder.Encode(make(chan int))
	require.Error(t, err)
	require.ErrorContains(t, err, "msgpack.Marshal")
}

func TestJSONCoder_roundtrip(t *testing.T) {
	coder := &JSONCoder{}
	original := testPayload{Name: "Bob", Age: 25, Email: "bob@example.com"}

	data, err := coder.Encode(original)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Bob","age":25,"email":"bob@example.com"}`, string(data))

	var decoded testPayload
	err = coder.Decode(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestJSONCoder_roundtrip_map(t *testing.T) {
	coder := &JSONCoder{}
	original := map[string]int{"x": 1, "y": 2}

	data, err := coder.Encode(original)
	require.NoError(t, err)

	var decoded map[string]int
	err = coder.Decode(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestJSONCoder_Decode_invalid(t *testing.T) {
	coder := &JSONCoder{}
	var dest testPayload
	err := coder.Decode([]byte("not json{{{"), &dest)
	require.Error(t, err)
	require.ErrorContains(t, err, "json.Unmarshal")
}

func TestJSONCoder_Encode_unsupported(t *testing.T) {
	coder := &JSONCoder{}
	_, err := coder.Encode(make(chan int))
	require.Error(t, err)
	require.ErrorContains(t, err, "json.Marshal")
}

// --- Benchmarks ---

type benchPayload struct {
	ID        int      `json:"id" msgpack:"id"`
	Name      string   `json:"name" msgpack:"name"`
	Email     string   `json:"email" msgpack:"email"`
	Age       int      `json:"age" msgpack:"age"`
	Active    bool     `json:"active" msgpack:"active"`
	Tags      []string `json:"tags" msgpack:"tags"`
	Score     float64  `json:"score" msgpack:"score"`
	Address   string   `json:"address" msgpack:"address"`
	CreatedAt int64    `json:"created_at" msgpack:"created_at"`
}

var benchData = benchPayload{
	ID:        42,
	Name:      "Alice Wonderland",
	Email:     "alice@example.com",
	Age:       30,
	Active:    true,
	Tags:      []string{"admin", "user", "premium"},
	Score:     98.6,
	Address:   "123 Main St, Springfield, IL 62701",
	CreatedAt: 1710000000,
}

func BenchmarkMsgPackCoder_Encode(b *testing.B) {
	coder := &MsgPackCoder{}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = coder.Encode(benchData)
	}
}

func BenchmarkMsgPackCoder_Decode(b *testing.B) {
	coder := &MsgPackCoder{}
	data, _ := coder.Encode(benchData)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var dest benchPayload
		_ = coder.Decode(data, &dest)
	}
}

func BenchmarkJSONCoder_Encode(b *testing.B) {
	coder := &JSONCoder{}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = coder.Encode(benchData)
	}
}

func BenchmarkJSONCoder_Decode(b *testing.B) {
	coder := &JSONCoder{}
	data, _ := coder.Encode(benchData)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var dest benchPayload
		_ = coder.Decode(data, &dest)
	}
}
