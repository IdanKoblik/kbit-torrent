package torrent

import (
	"testing"
	"kbit/pkg/types"
	"reflect"
)

func TestDecodeInt(t *testing.T) {
	input := "i42e"
	expected := types.BencodeInt(42)

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestDecodeString(t *testing.T) {
	input := "5:hello"
	expected := types.BencodeString("hello")

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestDecodeList(t *testing.T) {
	input := "l4:spam4:eggsi42ee"
	expected := types.BencodeList{
		types.BencodeString("spam"),
		types.BencodeString("eggs"),
		types.BencodeInt(42),
	}

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestDecodeDict(t *testing.T) {
	input := "d3:bar4:spam3:fooi42ee"
	expected := types.BencodeDict{
		"bar": types.BencodeString("spam"),
		"foo": types.BencodeInt(42),
	}

	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEncodeInt(t *testing.T) {
	input := types.BencodeInt(42)
	expected := "i42e"

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEncodeString(t *testing.T) {
	input := types.BencodeString("hello")
	expected := "5:hello"

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEncodeList(t *testing.T) {
	input := types.BencodeList{
		types.BencodeString("spam"),
		types.BencodeInt(42),
	}
	expected := "l4:spami42ee"

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEncodeDict(t *testing.T) {
	input := types.BencodeDict{
		"foo": types.BencodeInt(42),
		"bar": types.BencodeString("spam"),
	}

	expected1 := "d3:bar4:spam3:fooi42ee"
	expected2 := "d3:fooi42e3:bar4:spame"

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if result != expected1 && result != expected2 {
		t.Errorf("expected %v or %v, got %v", expected1, expected2, result)
	}
}

func TestRoundTrip(t *testing.T) {
	original := types.BencodeDict{
		"foo": types.BencodeList{
			types.BencodeInt(1),
			types.BencodeString("bar"),
		},
		"baz": types.BencodeString("spam"),
	}

	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("round-trip failed: expected %v, got %v", original, decoded)
	}
}

// --- Decode edge cases ---

func TestDecodeNegativeInt(t *testing.T) {
	result, err := Decode("i-42e")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result != types.BencodeInt(-42) {
		t.Errorf("expected -42, got %v", result)
	}
}

func TestDecodeZeroInt(t *testing.T) {
	result, err := Decode("i0e")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result != types.BencodeInt(0) {
		t.Errorf("expected 0, got %v", result)
	}
}

func TestDecodeEmptyString(t *testing.T) {
	result, err := Decode("0:")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result != types.BencodeString("") {
		t.Errorf("expected empty string, got %v", result)
	}
}

func TestDecodeEmptyList(t *testing.T) {
	result, err := Decode("le")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	list, ok := result.(types.BencodeList)
	if !ok {
		t.Fatalf("expected BencodeList, got %T", result)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %v", list)
	}
}

func TestDecodeEmptyDict(t *testing.T) {
	result, err := Decode("de")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	dict, ok := result.(types.BencodeDict)
	if !ok {
		t.Fatalf("expected BencodeDict, got %T", result)
	}
	if len(dict) != 0 {
		t.Errorf("expected empty dict, got %v", dict)
	}
}

func TestDecodeNestedStructure(t *testing.T) {
	// d3:keyl i1e i2e ee  →  {"key": [1, 2]}
	input := "d3:keyli1ei2eee"
	result, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	expected := types.BencodeDict{
		"key": types.BencodeList{
			types.BencodeInt(1),
			types.BencodeInt(2),
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestDecodeUnexpectedEndOfInput(t *testing.T) {
	_, err := Decode("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestDecodeUnknownType(t *testing.T) {
	_, err := Decode("x42")
	if err == nil {
		t.Error("expected error for unknown bencode type 'x'")
	}
}

func TestDecodeStringOutOfBounds(t *testing.T) {
	// String claims length 100 but data is much shorter.
	_, err := Decode("100:short")
	if err == nil {
		t.Error("expected error for out-of-bounds string")
	}
}

func TestDecodeInvalidStringFormat(t *testing.T) {
	_, err := Decode("abc")
	if err == nil {
		t.Error("expected error for string without colon separator")
	}
}

// --- Encode edge cases ---

func TestEncodeNegativeInt(t *testing.T) {
	result, err := Encode(types.BencodeInt(-100))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if result != "i-100e" {
		t.Errorf("expected i-100e, got %s", result)
	}
}

func TestEncodeEmptyString(t *testing.T) {
	result, err := Encode(types.BencodeString(""))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if result != "0:" {
		t.Errorf("expected 0:, got %s", result)
	}
}

func TestEncodeEmptyList(t *testing.T) {
	result, err := Encode(types.BencodeList{})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if result != "le" {
		t.Errorf("expected le, got %s", result)
	}
}

func TestEncodeEmptyDict(t *testing.T) {
	result, err := Encode(types.BencodeDict{})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if result != "de" {
		t.Errorf("expected de, got %s", result)
	}
}

func TestEncodeUnsupportedType(t *testing.T) {
	_, err := Encode(nil)
	if err == nil {
		t.Error("expected error for nil/unsupported type")
	}
}
