package torrent

import (
	"strconv"
	"unicode"
	"fmt"
	"kbit/pkg/types"
)

func Decode(str string) (types.BencodeValue, error) {
	pos := 0
	return decode(str, &pos)
}

func decode(str string, pos *int) (types.BencodeValue, error) {
	if *pos >= len(str) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	switch str[*pos] {
	case 'i':
		return decodeInt(str, pos)
	case 'l':
		return decodeList(str, pos)
	case 'd':
		return decodeDict(str, pos)
	default:
		if unicode.IsDigit(rune(str[*pos])) {
			return decodeString(str, pos)
		}
	}

	return nil, fmt.Errorf("unknown bencode type")
}

func decodeInt(str string, pos *int) (types.BencodeInt, error) {
	(*pos)++ // skip 'i'
	start := *pos

	for str[*pos] != 'e' {
		if *pos >= len(str) {
			return 0, fmt.Errorf("unterminated integer")
		}
		(*pos)++
	}

	val, err := strconv.ParseInt(str[start:*pos], 10, 64)
	if err != nil {
		return 0, err
	}

	(*pos)++ // skip 'e'
	return types.BencodeInt(val), nil
}

func decodeString(str string, pos *int) (types.BencodeString, error) {
	start := *pos

	for unicode.IsDigit(rune(str[*pos])) {
		(*pos)++
	}

	if str[*pos] != ':' {
		return "", fmt.Errorf("invalid string format")
	}

	length, err := strconv.Atoi(str[start:*pos])
	if err != nil {
		return "", err
	}

	(*pos)++ // skip ':'
	if *pos + length  > len(str) {
		return "", fmt.Errorf("string out of bounds")
	}

	result := str[*pos : *pos + length]
	*pos += length

	return types.BencodeString(result), nil
}

func decodeList(str string, pos *int) (types.BencodeList, error) {
	(*pos)++ // skip 'l'
	var list types.BencodeList

	for str[*pos] != 'e' {
		val, err := decode(str, pos)
		if err != nil {
			return nil, err
		}
		list = append(list, val)
	}

	(*pos)++ // skip 'e'
	return list, nil
}

func decodeDict(str string, pos *int) (types.BencodeDict, error) {
	(*pos)++ // skip 'd'
	dict := make(types.BencodeDict)

	for str[*pos] != 'e' {
		keyVal, err := decodeString(str, pos)
		if err != nil {
			return nil, err
		}

		val, err := decode(str, pos)
		if err != nil {
			return nil, err
		}

		dict[string(keyVal)] = val
	}

	(*pos)++ // skip 'e'
	return dict, nil
}

func Encode(v types.BencodeValue) (string, error) {
	switch val := v.(type) {
	case types.BencodeInt:
		return encodeInt(int64(val)), nil
	case types.BencodeString:
		return encodeString(string(val)), nil
	case types.BencodeList:
		return encodeList(val)
	case types.BencodeDict:
		return encodeDict(val)
	default:
		return "", fmt.Errorf("unsupported type %T", v)
	}
}

func encodeInt(i int64) string {
	return "i" + strconv.FormatInt(i, 10) + "e"
}

func encodeString(s string) string {
	return strconv.Itoa(len(s)) + ":" + s
}

func encodeList(list types.BencodeList) (string, error) {
	out := "l"
	for _, item := range list {
		enc, err := Encode(item)
		if err != nil {
			return "", err
		}
		out += enc
	}
	out += "e"
	return out, nil
}

func encodeDict(dict types.BencodeDict) (string, error) {
	out := "d"
	for key, value := range dict {
		out += encodeString(key)
		enc, err := Encode(value)
		if err != nil {
			return "", err
		}
		out += enc
	}
	out += "e"
	return out, nil
}
