package minhq_test

import (
	"bytes"
	"testing"

	"github.com/martinthomson/minhq"
	"github.com/stvp/assert"
)

type varintTest struct {
	v uint64
	e []byte
}

var varints = []varintTest{
	{0, []byte{0}},
	{1, []byte{1}},
	{63, []byte{63}},
	{64, []byte{0x40, 64}},
	{16383, []byte{0x7f, 0xff}},
	{16384, []byte{0x80, 0, 0x40, 0}},
	{1<<30 - 1, []byte{0xbf, 0xff, 0xff, 0xff}},
	{1 << 30, []byte{0xc0, 0, 0, 0, 0x40, 0, 0, 0}},
	{1<<62 - 1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
}

func TestVarintRead(t *testing.T) {
	for _, tc := range varints {
		fr := minhq.NewFrameReader(bytes.NewReader(tc.e))
		n, err := fr.ReadVarint()
		assert.Nil(t, err)
		assert.Equal(t, tc.v, n)
	}
}

func TestVarintReadLonger(t *testing.T) {
	var longerVarints = []varintTest{
		{0, []byte{0x40, 0}},
		{1, []byte{0x80, 0, 0, 1}},
		{63, []byte{0xc0, 0, 0, 0, 0, 0, 0, 63}},
		{64, []byte{0x80, 0, 0, 64}},
		{16383, []byte{0x80, 0, 0x3f, 0xff}},
		{16384, []byte{0xc0, 0, 0, 0, 0, 0, 0x40, 0}},
		{1<<30 - 1, []byte{0xc0, 0, 0, 0, 0x3f, 0xff, 0xff, 0xff}},
	}
	for _, tc := range longerVarints {
		fr := minhq.NewFrameReader(bytes.NewReader(tc.e))
		n, err := fr.ReadVarint()
		assert.Nil(t, err)
		assert.Equal(t, tc.v, n)
	}
}

func TestVarintWrite(t *testing.T) {
	for _, tc := range varints {
		var buf bytes.Buffer
		fw := minhq.NewFrameWriter(&buf)
		n, err := fw.WriteVarint(tc.v)
		assert.Nil(t, err)
		assert.Equal(t, len(tc.e), int(n))
		assert.Equal(t, tc.e, buf.Bytes())
	}
}

func TestVarintWriteOverflow(t *testing.T) {
	var buf bytes.Buffer
	fw := minhq.NewFrameWriter(&buf)
	_, err := fw.WriteVarint(1 << 63)
	assert.NotNil(t, err)
}

func TestFrameRead(t *testing.T) {
	fr := minhq.NewFrameReader(bytes.NewReader([]byte{1, 7, 0}))
	typ, r, err := fr.ReadFrame()
	assert.Nil(t, err)
	assert.Equal(t, minhq.FrameType(7), typ)
	var p [4]byte
	n, err := r.Read(p[:])
	assert.Nil(t, err)
	assert.Equal(t, []byte{0}, p[:n])
}
