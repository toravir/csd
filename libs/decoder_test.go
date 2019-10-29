package csd

import (
	"bytes"
	"encoding/hex"
	"net"
	"testing"
	"time"
)

func TestDecodeInteger(t *testing.T) {
	var integerTestCases = []struct {
		val    int
		binary string
	}{
		//value included in the type
		{0, "\x00"},
		{1, "\x01"},
		{2, "\x02"},
		{3, "\x03"},
		{8, "\x08"},
		{9, "\x09"},
		{10, "\x0a"},
		{22, "\x16"},
		{23, "\x17"},
		//Value in 1 byte
		{24, "\x18\x18"},
		{25, "\x18\x19"},
		{26, "\x18\x1a"},
		{100, "\x18\x64"},
		{254, "\x18\xfe"},
		{255, "\x18\xff"},
		//Value in 2 bytes
		{256, "\x19\x01\x00"},
		{257, "\x19\x01\x01"},
		{1000, "\x19\x03\xe8"},
		{0xFFFF, "\x19\xff\xff"},
		//Value in 4 bytes
		{0x10000, "\x1a\x00\x01\x00\x00"},
		{0xFFFFFFFE, "\x1a\xff\xff\xff\xfe"},
		{1000000, "\x1a\x00\x0f\x42\x40"},
		//Value in 8 bytes
		{0xabcd100000000, "\x1b\x00\x0a\xbc\xd1\x00\x00\x00\x00"},
		{1000000000000, "\x1b\x00\x00\x00\xe8\xd4\xa5\x10\x00"},
		// Negative number test cases
		//value included in the type
		{-1, "\x20"},
		{-2, "\x21"},
		{-3, "\x22"},
		{-10, "\x29"},
		{-21, "\x34"},
		{-22, "\x35"},
		{-23, "\x36"},
		{-24, "\x37"},
		//Value in 1 byte
		{-25, "\x38\x18"},
		{-26, "\x38\x19"},
		{-100, "\x38\x63"},
		{-254, "\x38\xfd"},
		{-255, "\x38\xfe"},
		{-256, "\x38\xff"},
		//Value in 2 bytes
		{-257, "\x39\x01\x00"},
		{-258, "\x39\x01\x01"},
		{-1000, "\x39\x03\xe7"},
		//Value in 4 bytes
		{-0x10001, "\x3a\x00\x01\x00\x00"},
		{-0xFFFFFFFE, "\x3a\xff\xff\xff\xfd"},
		{-1000000, "\x3a\x00\x0f\x42\x3f"},
		//Value in 8 bytes
		{-0xabcd100000001, "\x3b\x00\x0a\xbc\xd1\x00\x00\x00\x00"},
		{-1000000000001, "\x3b\x00\x00\x00\xe8\xd4\xa5\x10\x00"},
	}
	for _, tc := range integerTestCases {
		gotv := decodeInteger(getReader(tc.binary))
		if gotv != int64(tc.val) {
			t.Errorf("decodeInteger(0x%s)=0x%d, want: 0x%d",
				hex.EncodeToString([]byte(tc.binary)), gotv, tc.val)
		}
	}
}

func TestDecodeString(t *testing.T) {
	var encodeStringTests = []struct {
		plain  string
		binary string
		json   string //begin and end quotes are implied
	}{
		{"", "\x60", ""},
		{"\\", "\x61\x5c", "\\\\"},
		{"\x00", "\x61\x00", "\\u0000"},
		{"\x01", "\x61\x01", "\\u0001"},
		{"\x02", "\x61\x02", "\\u0002"},
		{"\x03", "\x61\x03", "\\u0003"},
		{"\x04", "\x61\x04", "\\u0004"},
		{"*", "\x61*", "*"},
		{"a", "\x61a", "a"},
		{"IETF", "\x64IETF", "IETF"},
		{"abcdefghijklmnopqrstuvwxyzABCD", "\x78\x1eabcdefghijklmnopqrstuvwxyzABCD", "abcdefghijklmnopqrstuvwxyzABCD"},
		{"<------------------------------------  This is a 100 character string ----------------------------->" +
			"<------------------------------------  This is a 100 character string ----------------------------->" +
			"<------------------------------------  This is a 100 character string ----------------------------->",
			"\x79\x01\x2c<------------------------------------  This is a 100 character string ----------------------------->" +
				"<------------------------------------  This is a 100 character string ----------------------------->" +
				"<------------------------------------  This is a 100 character string ----------------------------->",
			"<------------------------------------  This is a 100 character string ----------------------------->" +
				"<------------------------------------  This is a 100 character string ----------------------------->" +
				"<------------------------------------  This is a 100 character string ----------------------------->"},
		{"emoji \u2764\ufe0f!", "\x6demoji ❤️!", "emoji \u2764\ufe0f!"},
	}

	for _, tt := range encodeStringTests {
		got := decodeUTF8String(getReader(tt.binary))
		if string(got) != "\""+tt.json+"\"" {
			t.Errorf("DecodeString(0x%s)=%s, want:\"%s\"\n", hex.EncodeToString([]byte(tt.binary)), string(got),
				hex.EncodeToString([]byte(tt.json)))
		}
	}
}

func TestDecodeArray(t *testing.T) {
	var integerArrayTestCases = []struct {
		val    []int
		binary string
		json   string
	}{
		{[]int{-1, 0, 200, 20}, "\x84\x20\x00\x18\xc8\x14", "[-1,0,200,20]"},
		{[]int{-200, -10, 200, 400}, "\x84\x38\xc7\x29\x18\xc8\x19\x01\x90", "[-200,-10,200,400]"},
		{[]int{1, 2, 3}, "\x83\x01\x02\x03", "[1,2,3]"},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			"\x98\x19\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x18\x18\x19",
			"[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25]"},
	}
	for _, tc := range integerArrayTestCases {
		buf := bytes.NewBuffer([]byte{})
		array2Json(getReader(tc.binary), buf)
		if buf.String() != tc.json {
			t.Errorf("array2Json(0x%s)=%s, want: %s", hex.EncodeToString([]byte(tc.binary)), buf.String(), tc.json)
		}
	}
	//Unspecified Length Array
	var infiniteArrayTestCases = []struct {
		in  string
		out string
	}{
		{"\x9f\x20\x00\x18\xc8\x14\xff", "[-1,0,200,20]"},
		{"\x9f\x38\xc7\x29\x18\xc8\x19\x01\x90\xff", "[-200,-10,200,400]"},
		{"\x9f\x01\x02\x03\xff", "[1,2,3]"},
		{"\x9f\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x18\x18\x19\xff",
			"[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25]"},
	}
	for _, tc := range infiniteArrayTestCases {
		buf := bytes.NewBuffer([]byte{})
		array2Json(getReader(tc.in), buf)
		if buf.String() != tc.out {
			t.Errorf("array2Json(0x%s)=%s, want: %s", hex.EncodeToString([]byte(tc.out)), buf.String(), tc.out)
		}
	}

	var booleanArrayTestCases = []struct {
		val    []bool
		binary string
		json   string
	}{
		{[]bool{true, false, true}, "\x83\xf5\xf4\xf5", "[true,false,true]"},
		{[]bool{true, false, false, true, false, true}, "\x86\xf5\xf4\xf4\xf5\xf4\xf5", "[true,false,false,true,false,true]"},
	}
	for _, tc := range booleanArrayTestCases {
		buf := bytes.NewBuffer([]byte{})
		array2Json(getReader(tc.binary), buf)
		if buf.String() != tc.json {
			t.Errorf("array2Json(0x%s)=%s, want: %s", hex.EncodeToString([]byte(tc.binary)), buf.String(), tc.json)
		}
	}
	//TODO add cases for arrays of other types
}

var infiniteMapDecodeTestCases = []struct {
	bin  []byte
	json string
}{
	{[]byte("\xbf\x64IETF\x20\xff"), "{\"IETF\":-1}"},
	{[]byte("\xbf\x65Array\x84\x20\x00\x18\xc8\x14\xff"), "{\"Array\":[-1,0,200,20]}"},
}

var mapDecodeTestCases = []struct {
	bin  []byte
	json string
}{
	{[]byte("\xa1\x64IETF\x20"), "{\"IETF\":-1}"},
	{[]byte("\xa1\x65Array\x84\x20\x00\x18\xc8\x14"), "{\"Array\":[-1,0,200,20]}"},
}

func TestDecodeMap(t *testing.T) {
	for _, tc := range mapDecodeTestCases {
		buf := bytes.NewBuffer([]byte{})
		map2Json(getReader(string(tc.bin)), buf)
		if buf.String() != tc.json {
			t.Errorf("map2Json(0x%s)=%s, want: %s", hex.EncodeToString(tc.bin), buf.String(), tc.json)
		}
	}
	for _, tc := range infiniteMapDecodeTestCases {
		buf := bytes.NewBuffer([]byte{})
		map2Json(getReader(string(tc.bin)), buf)
		if buf.String() != tc.json {
			t.Errorf("map2Json(0x%s)=%s, want: %s", hex.EncodeToString(tc.bin), buf.String(), tc.json)
		}
	}
}

func TestDecodeBool(t *testing.T) {
	var booleanTestCases = []struct {
		val    bool
		binary string
		json   string
	}{
		{true, "\xf5", "true"},
		{false, "\xf4", "false"},
	}
	for _, tc := range booleanTestCases {
		got := decodeSimpleFloat(getReader(tc.binary))
		if string(got) != tc.json {
			t.Errorf("decodeSimpleFloat(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.binary)), string(got), tc.json)
		}
	}
}

func TestDecodeFloat(t *testing.T) {
	var float32TestCases = []struct {
		val    string
		binary string
	}{
		{"0",  "\xfa\x00\x00\x00\x00"},
		{"1", "\xfa\x3f\x80\x00\x00"},
		{"1.5", "\xfa\x3f\xc0\x00\x00"},
		{"65504", "\xfa\x47\x7f\xe0\x00"},
		{"-4", "\xfa\xc0\x80\x00\x00"},
		{"0.000061035156", "\xfa\x38\x80\x00\x00"},
	}

	for _, tc := range float32TestCases {
		got := decodeSimpleFloat(getReader(tc.binary))
		if string(got) != tc.val {
			t.Errorf("decodeFloat(0x%s)=%s, want:%s\n", hex.EncodeToString([]byte(tc.binary)), got, tc.val)
		}
	}
}

func TestDecodeTimestamp(t *testing.T) {
	var timeIntegerTestcases = []struct {
		txt    string
		binary string
		rfcStr string
	}{
		{"2013-02-03T19:54:00-08:00", "\xc1\x1a\x51\x0f\x30\xd8", "2013-02-04T03:54:00Z"},
		{"1950-02-03T19:54:00-08:00", "\xc1\x3a\x25\x71\x93\xa7", "1950-02-04T03:54:00Z"},
	}
	DecodeTimeZone, _ = time.LoadLocation("UTC")
	for _, tc := range timeIntegerTestcases {
		tm := decodeTagData(getReader(tc.binary))
		if string(tm) != "\""+tc.rfcStr+"\"" {
			t.Errorf("decodeFloat(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.binary)), tm, tc.rfcStr)
		}
	}
	var timeFloatTestcases = []struct {
		rfcStr string
		out    string
	}{
		{"2006-01-02T15:04:05.999999-08:00", "\xc1\xfb\x41\xd0\xee\x6c\x59\x7f\xff\xfc"},
		{"1956-01-02T15:04:05.999999-08:00", "\xc1\xfb\xc1\xba\x53\x81\x1a\x00\x00\x11"},
	}
	for _, tc := range timeFloatTestcases {
		tm := decodeTagData(getReader(tc.out))
		//Since we convert to float and back - it may be slightly off - so
		//we cannot check for exact equality instead, we'll check it is
		//very close to each other Less than a Microsecond (lets not yet do nanosec)

		got, _ := time.Parse(string(tm), string(tm))
		want, _ := time.Parse(tc.rfcStr, tc.rfcStr)
		if got.Sub(want) > time.Microsecond {
			t.Errorf("decodeFloat(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.out)), tm, tc.rfcStr)
		}
	}
}

func TestDecodeNetworkAddr(t *testing.T) {
	var ipAddrTestCases = []struct {
		ipaddr net.IP
		text   string // ASCII representation of ipaddr
		binary string // CBOR representation of ipaddr
	}{
		{net.IP{10, 0, 0, 1}, "\"10.0.0.1\"", "\xd9\x01\x04\x44\x0a\x00\x00\x01"},
		{net.IP{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x8a, 0x2e, 0x03, 0x70, 0x73, 0x34},
			"\"2001:db8:85a3::8a2e:370:7334\"",
			"\xd9\x01\x04\x50\x20\x01\x0d\xb8\x85\xa3\x00\x00\x00\x00\x8a\x2e\x03\x70\x73\x34"},
	}
	for _, tc := range ipAddrTestCases {
		d1 := decodeTagData(getReader(tc.binary))
		if string(d1) != tc.text {
			t.Errorf("decodeNetworkAddr(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.binary)), d1, tc.text)
		}
	}
}

func TestDecodeMACAddr(t *testing.T) {
	var macAddrTestCases = []struct {
		macaddr net.HardwareAddr
		text    string // ASCII representation of macaddr
		binary  string // CBOR representation of macaddr
	}{
		{net.HardwareAddr{0x12, 0x34, 0x56, 0x78, 0x90, 0xab}, "\"12:34:56:78:90:ab\"", "\xd9\x01\x04\x46\x12\x34\x56\x78\x90\xab"},
		{net.HardwareAddr{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3}, "\"20:01:0d:b8:85:a3\"", "\xd9\x01\x04\x46\x20\x01\x0d\xb8\x85\xa3"},
	}

	for _, tc := range macAddrTestCases {
		d1 := decodeTagData(getReader(tc.binary))
		if string(d1) != tc.text {
			t.Errorf("decodeNetworkAddr(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.binary)), d1, tc.text)
		}
	}
}

func TestDecodeIPPrefix(t *testing.T) {
	var IPPrefixTestCases = []struct {
		pfx    net.IPNet
		text   string // ASCII representation of pfx
		binary string // CBOR representation of pfx
	}{
		{net.IPNet{IP: net.IP{0, 0, 0, 0}, Mask: net.CIDRMask(0, 32)}, "\"0.0.0.0/0\"", "\xd9\x01\x05\xa1\x44\x00\x00\x00\x00\x00"},
		{net.IPNet{IP: net.IP{192, 168, 0, 100}, Mask: net.CIDRMask(24, 32)}, "\"192.168.0.100/24\"",
			"\xd9\x01\x05\xa1\x44\xc0\xa8\x00\x64\x18\x18"},
	}

	for _, tc := range IPPrefixTestCases {
		d1 := decodeTagData(getReader(tc.binary))
		if string(d1) != tc.text {
			t.Errorf("decodeIPPrefix(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.binary)), d1, tc.text)
		}
	}
}

var compositeCborTestCases = []struct {
	binary []byte
	json   string
}{
	{[]byte("\xbf\x64IETF\x20\x65Array\x9f\x20\x00\x18\xc8\x14\xff\xff"), "{\"IETF\":-1,\"Array\":[-1,0,200,20]}\n"},
	{[]byte("\xbf\x64IETF\x64YES!\x65Array\x9f\x20\x00\x18\xc8\x14\xff\xff"), "{\"IETF\":\"YES!\",\"Array\":[-1,0,200,20]}\n"},
        {[]byte("\xbf\x65level\x64info\x67Float32\xfa\x40\x4c\xcc\xcd\xff"), "{\"level\":\"info\",\"Float32\":3.2}\n"},
}

func TestDecodeCbor2Json(t *testing.T) {
	for _, tc := range compositeCborTestCases {
		buf := bytes.NewBuffer([]byte{})
		err := Cbor2JsonManyObjects(getReader(string(tc.binary)), buf)
		if buf.String() != tc.json {
			t.Errorf("cbor2JsonManyObjects(0x%s)=%s, want: %s", hex.EncodeToString(tc.binary), buf.String(), tc.json)
		}
		if err != nil {
			t.Errorf("cbor2JsonManyObjects(0x%s)=%s, want: %s", hex.EncodeToString(tc.binary), buf.String(), tc.json)
		}
	}
}

var negativeCborTestCases = []struct {
	binary []byte
	errStr string
}{
	{[]byte("\xb9\x64IETF\x20\x65Array\x9f\x20\x00\x18\xc8\x14"), "Tried to Read 18 Bytes.. But hit end of file"},
	{[]byte("\xbf\x64IETF\x20\x65Array\x9f\x20\x00\x18\xc8\x14"), "EOF"},
	{[]byte("\xbf\x14IETF\x20\x65Array\x9f\x20\x00\x18\xc8\x14"), "Tried to Read 40736 Bytes.. But hit end of file"},
	{[]byte("\xbf\x64IETF"), "EOF"},
	{[]byte("\xbf\x64IETF\x20\x65Array\x9f\x20\x00\x18\xc8\xff\xff\xff"), "Invalid Additional Type: 31 in decodeSimpleFloat"},
	{[]byte("\xbf\x64IETF\x20\x65Array"), "EOF"},
	{[]byte("\xbf\x64"), "Tried to Read 4 Bytes.. But hit end of file"},
}

func TestDecodeNegativeCbor2Json(t *testing.T) {
	for _, tc := range negativeCborTestCases {
		buf := bytes.NewBuffer([]byte{})
		err := Cbor2JsonManyObjects(getReader(string(tc.binary)), buf)
		if err == nil || err.Error() != tc.errStr {
			t.Errorf("Expected error got:%s, want:%s", err, tc.errStr)
		}
	}
}
