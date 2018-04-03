package tracer

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ugorji/go/codec"
)

// TestPayloadIntegrity tests that whatever we push into the payload
// allows us to read the same content as would have been encoded by
// the codec.
func TestPayloadIntegrity(t *testing.T) {
	assert := assert.New(t)
	p := newPayload()
	want := new(bytes.Buffer)
	for _, items := range [][]interface{}{
		{1, 2, 3},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
	} {
		p.reset()
		for _, v := range items {
			p.push(v)
		}
		assert.Equal(p.itemCount(), len(items))
		got, err := ioutil.ReadAll(p)
		assert.NoError(err)
		want.Reset()
		err = codec.NewEncoder(want, &codec.MsgpackHandle{}).Encode(items)
		assert.NoError(err)
		assert.Equal(want.Bytes(), got)
	}
}

// TestPayloadDecodeInts tests that whatever we push into the payload can
// be decoded by the codec.
func TestPayloadDecodeInts(t *testing.T) {
	assert := assert.New(t)
	p := newPayload()
	for _, items := range [][]int64{
		{1, 2, 3},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
	} {
		p.reset()
		for _, v := range items {
			p.push(v)
		}
		var got []int64
		err := codec.NewDecoder(p, &codec.MsgpackHandle{}).Decode(&got)
		assert.NoError(err)
		assert.Equal(items, got)
	}
}

// TestPayloadDecodetests that whatever we push into the payload can
// be decoded by the codec.
func TestPayloadDecode(t *testing.T) {
	assert := assert.New(t)
	p := newPayload()
	type AB struct{ A, B int }
	x := AB{1, 2}
	for _, items := range [][]AB{
		{x, x, x},
		{x, x, x, x, x, x, x, x, x, x, x, x, x, x},
		{x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x},
	} {
		p.reset()
		for _, v := range items {
			p.push(v)
		}
		var got []AB
		err := codec.NewDecoder(p, &codec.MsgpackHandle{}).Decode(&got)
		assert.NoError(err)
		assert.Equal(items, got)
	}
}

func BenchmarkPayloadThroughput10K(b *testing.B) {
	s := newBasicSpan("X")
	s.Meta["key"] = strings.Repeat("X", 10000)
	benchmarkPayloadThroughput(b, s)
}

func BenchmarkPayloadThroughput100K(b *testing.B) {
	s := newBasicSpan("X")
	s.Meta["key"] = strings.Repeat("X", 10000)
	trace := make([]*span, 10)
	for i := 0; i < 10; i++ {
		trace[i] = s
	}
	benchmarkPayloadThroughput(b, trace)
}

func BenchmarkPayloadThroughput1MB(b *testing.B) {
	s := newBasicSpan("X")
	s.Meta["key"] = strings.Repeat("X", 10000)
	trace := make([]*span, 100)
	for i := 0; i < 10; i++ {
		trace[i] = s
	}
	benchmarkPayloadThroughput(b, trace)
}

func benchmarkPayloadThroughput(b *testing.B, s interface{}) {
	p := newPayload()
	pkg := new(bytes.Buffer)
	if err := codec.NewEncoder(pkg, &codec.MsgpackHandle{}).Encode(s); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(pkg.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.push(s)
	}
}
