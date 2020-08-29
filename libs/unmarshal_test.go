package csd

import (
	//"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalString(t *testing.T) {
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
		got := unmarshalUTF8String(getReader(tt.binary))
		if string(got) != tt.plain {
			t.Errorf("UnmarshalString(0x%s)=%s, want:\"%s\"\n", hex.EncodeToString([]byte(tt.binary)), string(got),
				tt.plain)
		}
	}
}

func TestUnmarshalArray(t *testing.T) {
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
		got := unmarshalArray(getReader(tc.binary))
		if len(got) != len(tc.val) {
			t.Errorf("unmarshalArray(0x%s)=%v, want: %v", hex.EncodeToString([]byte(tc.binary)), got, tc.val)
		}
		for i := 0; i < len(tc.val); i++ {
			if i == len(got) {
				break
			}
			g := got[i].(int64)
			if int(g) != tc.val[i] {
				t.Errorf("unmarshalArray(0x%s)=%v, want: %v", hex.EncodeToString([]byte(tc.binary)), got, tc.val)
			}
		}
	}
	//Unspecified Length Array
	var infiniteArrayTestCases = []struct {
		in  string
		out []int
	}{
		{"\x9f\x20\x00\x18\xc8\x14\xff", []int{-1, 0, 200, 20}},
		{"\x9f\x38\xc7\x29\x18\xc8\x19\x01\x90\xff", []int{-200, -10, 200, 400}},
		{"\x9f\x01\x02\x03\xff", []int{1, 2, 3}},
		{"\x9f\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x18\x18\x19\xff",
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
	}
	for _, tc := range infiniteArrayTestCases {
		got := unmarshalArray(getReader(tc.in))
		if len(got) != len(tc.out) {
			t.Errorf("unmarshalArray(0x%s)=%v, want: %v", hex.EncodeToString([]byte(tc.in)), got, tc.out)
		}
		for i := 0; i < len(tc.out); i++ {
			if i == len(got) {
				break
			}
			g := got[i].(int64)
			if int(g) != tc.out[i] {
				t.Errorf("unmarshalArray(0x%s)=%v, want: %v", hex.EncodeToString([]byte(tc.in)), got, tc.out)
			}
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
		got := unmarshalArray(getReader(tc.binary))
		for i := 0; i < len(tc.val); i++ {
			if got[i].(bool) != tc.val[i] {
				t.Errorf("unmarshalArray(0x%s)=%v, want: %v", hex.EncodeToString([]byte(tc.binary)), got, tc.val)
			}
		}
	}

	//TODO add cases for arrays of other types
}

func TestUnmarshalBool(t *testing.T) {
	var booleanTestCases = []struct {
		val    bool
		binary string
		json   string
	}{
		{true, "\xf5", "true"},
		{false, "\xf4", "false"},
	}
	for _, tc := range booleanTestCases {
		got := unmarshalSimpleFloat(getReader(tc.binary))
		if got != tc.val {
			t.Errorf("unmarshalSimpleFloat(0x%s)=%v, want:%v", hex.EncodeToString([]byte(tc.binary)), got, tc.val)
		}
	}
}

func TestUnmarshalFloat(t *testing.T) {
	var float32TestCases = []struct {
		val    float64
		binary string
	}{
		{0, "\xfa\x00\x00\x00\x00"},
		{1, "\xfa\x3f\x80\x00\x00"},
		{1.5, "\xfa\x3f\xc0\x00\x00"},
		{65504, "\xfa\x47\x7f\xe0\x00"},
		{-4, "\xfa\xc0\x80\x00\x00"},
		{0.000061035156, "\xfa\x38\x80\x00\x00"},
	}

	for _, tc := range float32TestCases {
		got := unmarshalSimpleFloat(getReader(tc.binary))
		g := got.(float64)
		if g != tc.val && (g-tc.val > 0.000001 || g-tc.val < -0.000001) {
			t.Errorf("unmarshalFloat(0x%s)=%v, want:%v delta:%v\n", hex.EncodeToString([]byte(tc.binary)), got, tc.val, g-tc.val)
		}
	}
}

func isSameIpAddr(p1, p2 net.IP) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i, v := range p1 {
		if v != p2[i] {
			return false
		}
	}
	return true
}

func TestUnmarshalNetworkAddr(t *testing.T) {
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
		d1 := unmarshalTagData(getReader(tc.binary))
		if !isSameIpAddr(d1.(net.IP), tc.ipaddr) {
			t.Errorf("unmarshalNetworkAddr(0x%s)=%v, want:%v", hex.EncodeToString([]byte(tc.binary)), d1, tc.ipaddr)
		}
	}
}

func isSameMacAddr(p1, p2 net.HardwareAddr) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i, v := range p1 {
		if v != p2[i] {
			return false
		}
	}
	return true
}

func TestUnmarshalMACAddr(t *testing.T) {
	var macAddrTestCases = []struct {
		macaddr net.HardwareAddr
		text    string // ASCII representation of macaddr
		binary  string // CBOR representation of macaddr
	}{
		{net.HardwareAddr{0x12, 0x34, 0x56, 0x78, 0x90, 0xab}, "\"12:34:56:78:90:ab\"", "\xd9\x01\x04\x46\x12\x34\x56\x78\x90\xab"},
		{net.HardwareAddr{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3}, "\"20:01:0d:b8:85:a3\"", "\xd9\x01\x04\x46\x20\x01\x0d\xb8\x85\xa3"},
	}

	for _, tc := range macAddrTestCases {
		d1 := unmarshalTagData(getReader(tc.binary))
		if !isSameMacAddr(d1.(net.HardwareAddr), tc.macaddr) {
			t.Errorf("unmarshalNetworkAddr(0x%s)=%v, want:%v", hex.EncodeToString([]byte(tc.binary)), d1, tc.macaddr)
		}
	}
}

func isSameIpPrefix(p1, p2 net.IPNet) bool {
	m1, l1 := p1.Mask.Size()
	m2, l2 := p2.Mask.Size()
	return p1.IP.Equal(p2.IP) && m1 == m2 && l1 == l2
}

func TestUnmarshalIPPrefix(t *testing.T) {
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
		d1 := unmarshalTagData(getReader(tc.binary))
		if !isSameIpPrefix(d1.(net.IPNet), tc.pfx) {
			t.Errorf("unmarshalIPPrefix(0x%s)=%v, want:%v", hex.EncodeToString([]byte(tc.binary)), d1, tc.pfx)
		}
	}
}

func TestUnmarshalTimestamp(t *testing.T) {
	var timeIntegerTestcases = []struct {
		binary string
		rfcStr string
	}{
		{"\xc1\x1a\x51\x0f\x30\xd8", "2013-02-04T03:54:00Z"},
		{"\xc1\x3a\x25\x71\x93\xa7", "1950-02-04T03:54:00Z"},
	}
	for _, tc := range timeIntegerTestcases {
		tm := unmarshalTagData(getReader(tc.binary))
		want, e := time.Parse(time.RFC3339, tc.rfcStr)
		if e != nil {
			fmt.Println(e)
		}
		got := tm.(time.Time)
		if got != want {
			t.Errorf("unmarshalFloat(0x%s)=%v, want:%v", hex.EncodeToString([]byte(tc.binary)), tm, want)
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
		tm := unmarshalTagData(getReader(tc.out))
		//Since we convert to float and back - it may be slightly off - so
		//we cannot check for exact equality instead, we'll check it is
		//very close to each other Less than a Microsecond (lets not yet do nanosec)

		got := tm.(time.Time)
		want, _ := time.Parse(time.RFC3339, tc.rfcStr)
		if got.Sub(want) > time.Microsecond {
			t.Errorf("unmarshalFloat(0x%s)=%s, want:%s", hex.EncodeToString([]byte(tc.out)), tm, tc.rfcStr)
		}
	}
}

func deRef(p reflect.Value) reflect.Value {
	if p.Kind() == reflect.Interface {
		p = p.Elem()
	}
	return p
}

func isSliceEqual(va1, va2 reflect.Value) bool {
	if va1.Len() != va2.Len() {
		return false
	}
	for i := 0; i < va1.Len(); i++ {
		v1 := deRef(va1.Index(i))
		v2 := deRef(va2.Index(i))
		if v1.Kind() != v2.Kind() {
			return false
		}
		switch v1.Kind() {
		case reflect.Int64:
			i1 := v1.Int()
			i2 := v2.Int()
			if i1 != i2 {
				return false
			}
		default:
			fmt.Println("Skipping -", v1.Kind())
		}
	}
	return true
}

func isMapSame(m1, m2 map[string]interface{}) bool {
	if len(m1) != len(m2) {
		return false
	}
	for ke1, v1 := range m1 {
		v2, ok := m2[ke1]
		if !ok {
			return false
		}
		ki1 := reflect.TypeOf(v1).Kind()
		ki2 := reflect.TypeOf(v2).Kind()
		if ki1 != ki2 {
			return false
		}
		va1, va2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
		switch ki1 {
		case reflect.Bool:
			if va1.Bool() != va2.Bool() {
				return false
			}
		case reflect.Int, reflect.Int64:
			if va1.Int() != va2.Int() {
				return false
			}
		case reflect.Array, reflect.Slice:
			if !isSliceEqual(reflect.ValueOf(v1), reflect.ValueOf(v2)) {
				return false
			}
		case reflect.String:
			if va1.String() != va2.String() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			d := va1.Float() - va2.Float()
			if d < -0.001 || d > 0.001 {
				return false
			}
		default:
			fmt.Println("Skipping -", ki1)
		}
	}
	return true
}

var infiniteMapUnmarshalTestCases = []struct {
	bin  []byte
	want map[string]interface{}
}{
	{[]byte("\xbf\x64IETF\x20\xff"), map[string]interface{}{"IETF": int64(-1)}},
	{[]byte("\xbf\x65Array\x84\x20\x00\x18\xc8\x14\xff"), map[string]interface{}{"Array": []int64{int64(-1), int64(0), int64(200), int64(20)}}},
}

var mapUnmarshalTestCases = []struct {
	bin  []byte
	want map[string]interface{}
}{
	{[]byte("\xa1\x64IETF\x20"), map[string]interface{}{"IETF": int64(-1)}},
	{[]byte("\xa1\x65Array\x84\x20\x00\x18\xc8\x14"), map[string]interface{}{"Array": []int64{int64(-1), int64(0), int64(200), int64(20)}}},
}

func TestUnmarshalMap(t *testing.T) {
	for _, tc := range mapUnmarshalTestCases {
		got := unmarshalMap(getReader(string(tc.bin)))
		if !isMapSame(got, tc.want) {
			t.Errorf("unmarshalMap(0x%s)=%v, want: %v", hex.EncodeToString(tc.bin), got, tc.want)
		}
	}
	for _, tc := range infiniteMapUnmarshalTestCases {
		got := unmarshalMap(getReader(string(tc.bin)))
		if !isMapSame(got, tc.want) {
			t.Errorf("unmarshalMap(0x%s)=%v, want: %v", hex.EncodeToString(tc.bin), got, tc.want)
		}
	}
}

var compositeCborUnmarshalTestCases = []struct {
	binary []byte
	want   map[string]interface{}
}{
	{[]byte("\xbf\x64IETF\x20\x65Array\x9f\x20\x00\x18\xc8\x14\xff\xff"), map[string]interface{}{"IETF": int64(-1), "Array": []int64{-1, 0, 200, 20}}},
	{[]byte("\xbf\x64IETF\x64YES!\x65Array\x9f\x20\x00\x18\xc8\x14\xff\xff"), map[string]interface{}{"IETF": "YES!", "Array": []int64{-1, 0, 200, 20}}},
	{[]byte("\xbf\x65level\x64info\x67Float32\xfa\x40\x4c\xcc\xcd\xff"), map[string]interface{}{"level": "info", "Float32": 3.2}},
}

func TestUnmarshalCbor2Json(t *testing.T) {
	for _, tc := range compositeCborUnmarshalTestCases {
		got := unmarshalMap(getReader(string(tc.binary)))
		if !isMapSame(got, tc.want) {
			t.Errorf("cbor2JsonManyObjects(0x%s)=%v, want: %v", hex.EncodeToString(tc.binary), got, tc.want)
		}
	}
}
