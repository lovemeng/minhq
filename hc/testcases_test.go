package hc_test

import (
	"testing"

	"github.com/martinthomson/minhq/hc"
	"github.com/stvp/assert"
)

type dynamicTableEntry struct {
	name  string
	value string
}

func checkDynamicTable(t *testing.T, table *hc.Table, entries []dynamicTableEntry) {
	for i, e := range entries {
		// The initial offset for dynamic entries is 62 in HPACK.
		entry := table.Get(i + 62)
		assert.NotNil(t, entry)
		assert.Equal(t, e.name, entry.Name())
		assert.Equal(t, e.value, entry.Value())
	}
}

var testCases = []struct {
	resetTable   bool
	headers      []hc.HeaderField
	huffman      bool
	hpack        string
	qcramControl string
	qcramHeader  string
	tableSize    hc.TableCapacity
	dynamicTable []dynamicTableEntry
}{
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: "custom-key", Value: "custom-header", Sensitive: false},
		},
		huffman:      false,
		hpack:        "400a637573746f6d2d6b65790d637573746f6d2d686561646572",
		qcramControl: "400a637573746f6d2d6b65790d637573746f6d2d686561646572",
		qcramHeader:  "01be",
		tableSize:    55,
		dynamicTable: []dynamicTableEntry{
			{"custom-key", "custom-header"},
		},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":path", Value: "/sample/path", Sensitive: false},
		},
		huffman:      false,
		hpack:        "040c2f73616d706c652f70617468",
		qcramControl: "",
		qcramHeader:  "00040c2f73616d706c652f70617468",
		tableSize:    0,
		dynamicTable: []dynamicTableEntry{},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: "password", Value: "secret", Sensitive: true},
		},
		huffman:      false,
		hpack:        "100870617373776f726406736563726574",
		qcramControl: "",
		qcramHeader:  "00100870617373776f726406736563726574",
		tableSize:    0,
		dynamicTable: []dynamicTableEntry{},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
		},
		huffman:      false,
		hpack:        "82",
		qcramControl: "",
		qcramHeader:  "0082",
		tableSize:    0,
		dynamicTable: []dynamicTableEntry{},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "http", Sensitive: false},
			{Name: ":path", Value: "/", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
		},
		huffman:      false,
		hpack:        "828684410f7777772e6578616d706c652e636f6d",
		qcramControl: "410f7777772e6578616d706c652e636f6d",
		qcramHeader:  "01828684be",
		tableSize:    57,
		dynamicTable: []dynamicTableEntry{
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "http", Sensitive: false},
			{Name: ":path", Value: "/", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
			{Name: "cache-control", Value: "no-cache", Sensitive: false},
		},
		huffman:      false,
		hpack:        "828684be58086e6f2d6361636865",
		qcramControl: "58086e6f2d6361636865",
		qcramHeader:  "02828684bfbe",
		tableSize:    110,
		dynamicTable: []dynamicTableEntry{
			{"cache-control", "no-cache"},
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "https", Sensitive: false},
			{Name: ":path", Value: "/index.html", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
			{Name: "custom-key", Value: "custom-value", Sensitive: false},
		},
		huffman:      false,
		hpack:        "828785bf400a637573746f6d2d6b65790c637573746f6d2d76616c7565",
		qcramControl: "400a637573746f6d2d6b65790c637573746f6d2d76616c7565",
		qcramHeader:  "03828785c0be",
		tableSize:    164,
		dynamicTable: []dynamicTableEntry{
			{"custom-key", "custom-value"},
			{"cache-control", "no-cache"},
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "http", Sensitive: false},
			{Name: ":path", Value: "/", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
		},
		huffman:      true,
		hpack:        "828684418cf1e3c2e5f23a6ba0ab90f4ff",
		qcramControl: "418cf1e3c2e5f23a6ba0ab90f4ff",
		qcramHeader:  "01828684be",
		tableSize:    57,
		dynamicTable: []dynamicTableEntry{
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "http", Sensitive: false},
			{Name: ":path", Value: "/", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
			{Name: "cache-control", Value: "no-cache", Sensitive: false},
		},
		huffman:      true,
		hpack:        "828684be5886a8eb10649cbf",
		qcramControl: "5886a8eb10649cbf",
		qcramHeader:  "02828684bfbe",
		tableSize:    110,
		dynamicTable: []dynamicTableEntry{
			{"cache-control", "no-cache"},
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":method", Value: "GET", Sensitive: false},
			{Name: ":scheme", Value: "https", Sensitive: false},
			{Name: ":path", Value: "/index.html", Sensitive: false},
			{Name: ":authority", Value: "www.example.com", Sensitive: false},
			{Name: "custom-key", Value: "custom-value", Sensitive: false},
		},
		huffman:      true,
		hpack:        "828785bf408825a849e95ba97d7f8925a849e95bb8e8b4bf",
		qcramControl: "408825a849e95ba97d7f8925a849e95bb8e8b4bf",
		qcramHeader:  "03828785c0be",
		tableSize:    164,
		dynamicTable: []dynamicTableEntry{
			{"custom-key", "custom-value"},
			{"cache-control", "no-cache"},
			{":authority", "www.example.com"},
		},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "302", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:21 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
		},
		huffman: false,
		hpack: "4803333032580770726976617465611d4d6f6e2c203231204f63742032303133" +
			"2032303a31333a323120474d546e1768747470733a2f2f7777772e6578616d70" +
			"6c652e636f6d",
		qcramControl: "4803333032580770726976617465611d4d6f6e2c203231204f63742032303133" +
			"2032303a31333a323120474d546e1768747470733a2f2f7777772e6578616d70" +
			"6c652e636f6d",
		qcramHeader: "04c1c0bfbe",
		tableSize:   222,
		dynamicTable: []dynamicTableEntry{
			{"location", "https://www.example.com"},
			{"date", "Mon, 21 Oct 2013 20:13:21 GMT"},
			{"cache-control", "private"},
			{":status", "302"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "307", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:21 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
		},
		huffman:      false,
		hpack:        "4803333037c1c0bf",
		qcramControl: "4803333037",
		qcramHeader:  "05bec1c0bf",
		tableSize:    222,
		dynamicTable: []dynamicTableEntry{
			{":status", "307"},
			{"location", "https://www.example.com"},
			{"date", "Mon, 21 Oct 2013 20:13:21 GMT"},
			{"cache-control", "private"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "200", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:22 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
			{Name: "content-encoding", Value: "gzip", Sensitive: false},
			{Name: "set-cookie",
				Value:     "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1",
				Sensitive: false},
		},
		huffman: false,
		hpack: "88c1611d4d6f6e2c203231204f637420323031332032303a31333a323220474d" +
			"54c05a04677a69707738666f6f3d4153444a4b48514b425a584f5157454f5049" +
			"5541585157454f49553b206d61782d6167653d333630303b2076657273696f6e" +
			"3d31",
		qcramControl: "611d4d6f6e2c203231204f637420323031332032303a31333a323220474d" +
			"545a04677a69707738666f6f3d4153444a4b48514b425a584f5157454f5049" +
			"5541585157454f49553b206d61782d6167653d333630303b2076657273696f6e3d31",
		qcramHeader: "0888c4c0c2bfbe",
		tableSize:   215,
		dynamicTable: []dynamicTableEntry{
			{"set-cookie", "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1"},
			{"content-encoding", "gzip"},
			{"date", "Mon, 21 Oct 2013 20:13:22 GMT"},
		},
	},
	{
		resetTable: true,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "302", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:21 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
		},
		huffman: true,
		hpack: "488264025885aec3771a4b6196d07abe941054d444a8200595040b8166e082a6" +
			"2d1bff6e919d29ad171863c78f0b97c8e9ae82ae43d3",
		qcramControl: "488264025885aec3771a4b6196d07abe941054d444a8200595040b8166e082a6" +
			"2d1bff6e919d29ad171863c78f0b97c8e9ae82ae43d3",
		qcramHeader: "04c1c0bfbe",
		tableSize:   222,
		dynamicTable: []dynamicTableEntry{
			{"location", "https://www.example.com"},
			{"date", "Mon, 21 Oct 2013 20:13:21 GMT"},
			{"cache-control", "private"},
			{":status", "302"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "307", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:21 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
		},
		huffman:      true,
		hpack:        "4883640effc1c0bf",
		qcramControl: "4883640eff",
		qcramHeader:  "05bec1c0bf",
		tableSize:    222,
		dynamicTable: []dynamicTableEntry{
			{":status", "307"},
			{"location", "https://www.example.com"},
			{"date", "Mon, 21 Oct 2013 20:13:21 GMT"},
			{"cache-control", "private"},
		},
	},
	{
		resetTable: false,
		headers: []hc.HeaderField{
			{Name: ":status", Value: "200", Sensitive: false},
			{Name: "cache-control", Value: "private", Sensitive: false},
			{Name: "date", Value: "Mon, 21 Oct 2013 20:13:22 GMT", Sensitive: false},
			{Name: "location", Value: "https://www.example.com", Sensitive: false},
			{Name: "content-encoding", Value: "gzip", Sensitive: false},
			{Name: "set-cookie",
				Value:     "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1",
				Sensitive: false},
		},
		huffman: true,
		hpack: "88c16196d07abe941054d444a8200595040b8166e084a62d1bffc05a839bd9ab" +
			"77ad94e7821dd7f2e6c7b335dfdfcd5b3960d5af27087f3672c1ab270fb5291f" +
			"9587316065c003ed4ee5b1063d5007",
		qcramControl: "6196d07abe941054d444a8200595040b8166e084a62d1bff5a839bd9ab" +
			"77ad94e7821dd7f2e6c7b335dfdfcd5b3960d5af27087f3672c1ab270fb5291f" +
			"9587316065c003ed4ee5b1063d5007",
		qcramHeader: "0888c4c0c2bfbe",
		tableSize:   215,
		dynamicTable: []dynamicTableEntry{
			{"set-cookie", "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1"},
			{"content-encoding", "gzip"},
			{"date", "Mon, 21 Oct 2013 20:13:22 GMT"},
		},
	},
}
