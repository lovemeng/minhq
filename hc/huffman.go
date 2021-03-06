package hc

import (
	"errors"
	"io"

	bitio "github.com/martinthomson/minhq/io"
)

// HuffmanCompressor is a progressive compressor for Huffman-encoded data.
type HuffmanCompressor struct {
	writer    bitio.BitWriter
	saved     byte
	savedBits byte
}

// NewHuffmanCompressor wraps the underlying io.Writer.
func NewHuffmanCompressor(writer io.Writer) *HuffmanCompressor {
	return &HuffmanCompressor{bitio.NewBitWriter(writer), 0, 0}
}

// Add compresses a string using the Huffman table.  Strings are provided as byte slices.
func (compressor *HuffmanCompressor) Write(input []byte) (int, error) {
	for i, c := range input {
		entry := huffmanTable[c]
		err := compressor.writer.WriteBits(uint64(entry.val), entry.len)
		if err != nil {
			return i, err
		}
	}
	return len(input), nil
}

// Pad adds a terminator value and returns the full compressed value.
func (compressor *HuffmanCompressor) Pad() error {
	return compressor.writer.Pad(0xff)
}

// This is a huffmanDecoderNode in the reverse mapping tree.  We use 4-bit chunks because those result in at most a single emission of a character.
type huffmanDecoderNode struct {
	next [2]*huffmanDecoderNode
	leaf bool
	val  byte
}

func makeLayer(prefix uint32, prefixLen byte) *huffmanDecoderNode {
	layer := new(huffmanDecoderNode)
	found := false
	for i, e := range huffmanTable {
		if e.len < prefixLen+1 {
			continue
		}
		if (e.val >> (e.len - prefixLen)) != prefix {
			continue
		}
		arity := (e.val >> (e.len - prefixLen - 1)) & 1
		var child *huffmanDecoderNode
		if e.len == prefixLen+1 {
			child = new(huffmanDecoderNode)
			child.leaf = true
			child.val = byte(i)
			layer.next[arity] = child
			if layer.next[arity^1] != nil {
				return layer
			}
		}
		found = true
	}
	// There are unused parts of the tree, so leave the branches as nil if we reach those
	if found {
		if layer.next[0] == nil {
			layer.next[0] = makeLayer(prefix<<1, prefixLen+1)
		}
		if layer.next[1] == nil {
			layer.next[1] = makeLayer((prefix<<1)|1, prefixLen+1)
		}
	}
	return layer
}

var decompressorTree *huffmanDecoderNode

func initDecompressorTree() {
	if decompressorTree == nil {
		decompressorTree = makeLayer(0, 0)
	}
}

// HuffmanDecompressor is the opposite of huffmanCompressor
type HuffmanDecompressor struct {
	reader bitio.BitReader
	cursor *huffmanDecoderNode
}

// NewHuffmanDecompressor makes a new decompressor, which implements io.Reader.
func NewHuffmanDecompressor(reader io.Reader) *HuffmanDecompressor {
	initDecompressorTree()
	return &HuffmanDecompressor{bitio.NewBitReader(reader), decompressorTree}
}

// Read bytes of input and decode.
func (decompressor *HuffmanDecompressor) Read(p []byte) (int, error) {
	i := 0
	for i < len(p) {
		b, err := decompressor.reader.ReadBit()
		if err != nil {
			return i, err
		}

		decompressor.cursor = decompressor.cursor.next[b]
		if decompressor.cursor == nil {
			return i, errors.New("invalid Huffman coding")
		}
		if decompressor.cursor.leaf {
			p[i] = decompressor.cursor.val
			i++
			decompressor.cursor = decompressorTree
		}
	}
	return i, nil
}
