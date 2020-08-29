package csd

// This file contains code to decode a stream of CBOR Data into a map[string]interface{}

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func unmarshalString(src *bufio.Reader, noQuotes bool) string {
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeByteString && major != majorTypeUtf8String {
		panic(fmt.Errorf("Major type is: %d in unmarshalString", major))
	}
	result := []byte{}
	if !noQuotes {
		result = append(result, '"')
	}
	length := decodeIntAdditonalType(src, minor)
	len := int(length)
	pbs := readNBytes(src, len)
	result = append(result, pbs...)
	if noQuotes {
		return string(result)
	}
	return string(append(result, '"'))
}

func unmarshalUTF8String(src *bufio.Reader) string {
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeUtf8String {
		panic(fmt.Errorf("Major type is: %d in decodeUTF8String", major))
	}
	result := []byte{}
	length := decodeIntAdditonalType(src, minor)
	len := int(length)
	pbs := readNBytes(src, len)
	result = append(result, pbs...)
	return string(result)
}

func unmarshalArray(src *bufio.Reader) []interface{} {
	ret := []interface{}{}
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeArray {
		panic(fmt.Errorf("Major type is: %d in array2Json", major))
	}
	len := 0
	unSpecifiedCount := false
	if minor == additionalTypeInfiniteCount {
		unSpecifiedCount = true
	} else {
		length := decodeIntAdditonalType(src, minor)
		len = int(length)
	}
	for i := 0; unSpecifiedCount || i < len; i++ {
		if unSpecifiedCount {
			pb, e := src.Peek(1)
			if e != nil {
				panic(e)
			}
			if pb[0] == byte(majorTypeSimpleAndFloat|additionalTypeBreak) {
				readByte(src)
				break
			}
		}
		ret = append(ret, unmarshalOneObject(src))
	}
	return ret
}

func unmarshalMap(src *bufio.Reader) map[string]interface{} {
	ret := make(map[string]interface{})
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeMap {
		panic(fmt.Errorf("Major type is: %d in map2Json", major))
	}
	len := 0
	unSpecifiedCount := false
	if minor == additionalTypeInfiniteCount {
		unSpecifiedCount = true
	} else {
		length := decodeIntAdditonalType(src, minor)
		len = int(length) * 2
	}
	k := ""
	for i := 0; unSpecifiedCount || i < len; i++ {
		if unSpecifiedCount {
			pb, e := src.Peek(1)
			if e != nil {
				panic(e)
			}
			if pb[0] == byte(majorTypeSimpleAndFloat|additionalTypeBreak) {
				readByte(src)
				break
			}
		}
		if i%2 == 0 {
			// Even position values are keys.
			k = unmarshalString(src, true)
		} else {
			v := unmarshalOneObject(src)
			ret[k] = v
		}
	}
	return ret
}

func unmarshalTagData(src *bufio.Reader) interface{} {
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeTags {
		panic(fmt.Errorf("Major type is: %d in decodeTagData", major))
	}
	switch minor {
	case additionalTypeTimestamp:
		return unmarshalTimeStamp(src)

	// Tag value is larger than 256 (so uint16).
	case additionalTypeIntUint16:
		val := decodeIntAdditonalType(src, minor)

		switch uint16(val) {
		case additionalTypeEmbeddedJSON:
			pb := readByte(src)
			dataMajor := pb & maskOutAdditionalType
			if dataMajor != majorTypeByteString {
				panic(fmt.Errorf("Unsupported embedded Type: %d in decodeEmbeddedJSON", dataMajor))
			}
			src.UnreadByte()
			s := unmarshalString(src, true)
			m := make(map[string]interface{})
			err := json.Unmarshal([]byte(s), m)
			if err != nil {
				panic(err)
			}
			return m

		case additionalTypeTagNetworkAddr:
			octets := decodeString(src, true)
			switch len(octets) {
			case 6: // MAC address.
				ha := net.HardwareAddr(octets)
				return ha
			case 4: // IPv4 address.
				fallthrough
			case 16: // IPv6 address.
				ip := net.IP(octets)
				return ip
			default:
				panic(fmt.Errorf("Unexpected Network Address length: %d (expected 4,6,16)", len(octets)))
			}

		case additionalTypeTagNetworkPrefix:
			pb := readByte(src)
			if pb != byte(majorTypeMap|0x1) {
				panic(fmt.Errorf("IP Prefix is NOT of MAP of 1 elements as expected"))
			}
			octets := decodeString(src, true)
			val := decodeInteger(src)
			ip := net.IP(octets)
			var mask net.IPMask
			pfxLen := int(val)
			if len(octets) == 4 {
				mask = net.CIDRMask(pfxLen, 32)
			} else {
				mask = net.CIDRMask(pfxLen, 128)
			}
			ipPfx := net.IPNet{IP: ip, Mask: mask}
			return ipPfx

		case additionalTypeTagHexString:
			octets := decodeString(src, true)
			ss := []byte{'"'}
			for _, v := range octets {
				ss = append(ss, hexTable[v>>4], hexTable[v&0x0f])
			}
			return append(ss, '"')

		default:
			panic(fmt.Errorf("Unsupported Additional Tag Type: %d in decodeTagData", val))
		}
	}
	panic(fmt.Errorf("Unsupported Additional Type: %d in decodeTagData", minor))
}

func unmarshalTimeStamp(src *bufio.Reader) interface{} {
	pb := readByte(src)
	src.UnreadByte()
	tsMajor := pb & maskOutAdditionalType
	if tsMajor == majorTypeUnsignedInt || tsMajor == majorTypeNegativeInt {
		n := decodeInteger(src)
		t := time.Unix(n, 0)
		t = t.In(time.UTC)
		return t
	} else if tsMajor == majorTypeSimpleAndFloat {
		n, _ := decodeFloat(src)
		secs := int64(n)
		n -= float64(secs)
		n *= float64(1e9)
		t := time.Unix(secs, int64(n))
		t = t.In(time.UTC)
		return t
	}
	panic(fmt.Errorf("TS format is neigther int nor float: %d", tsMajor))
}

func unmarshalSimpleFloat(src *bufio.Reader) interface{} {
	pb := readByte(src)
	major := pb & maskOutAdditionalType
	minor := pb & maskOutMajorType
	if major != majorTypeSimpleAndFloat {
		panic(fmt.Errorf("Major type is: %d in decodeSimpleFloat", major))
	}
	switch minor {
	case additionalTypeBoolTrue:
		return true
	case additionalTypeBoolFalse:
		return false
	case additionalTypeNull:
		return nil
	case additionalTypeFloat16:
		fallthrough
	case additionalTypeFloat32:
		fallthrough
	case additionalTypeFloat64:
		src.UnreadByte()
		v, _ := decodeFloat(src)
		return v
	default:
		panic(fmt.Errorf("Invalid Additional Type: %d in decodeSimpleFloat", minor))
	}
}

func unmarshalOneObject(src *bufio.Reader) interface{} {
	pb, e := src.Peek(1)
	if e != nil {
		panic(e)
	}
	major := (pb[0] & maskOutAdditionalType)

	switch major {
	case majorTypeUnsignedInt:
		fallthrough
	case majorTypeNegativeInt:
		n := decodeInteger(src)
		return n

	case majorTypeByteString:
		s := decodeString(src, true)
		return s

	case majorTypeUtf8String:
		s := unmarshalUTF8String(src)
		return s

	case majorTypeArray:
		return unmarshalArray(src)

	case majorTypeMap:
		return unmarshalMap(src)

	case majorTypeTags:
		s := unmarshalTagData(src)
		return s

	case majorTypeSimpleAndFloat:
		s := unmarshalSimpleFloat(src)
		return s
	}
	var v interface{}
	return v
}
