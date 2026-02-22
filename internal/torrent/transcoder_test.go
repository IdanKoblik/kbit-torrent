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
