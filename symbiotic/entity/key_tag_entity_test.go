package entity

import (
	"strings"
	"testing"
)

func TestKeyTag_Type(t *testing.T) {
	cases := []struct {
		input    KeyTag
		expected KeyType
	}{
		{0x00, KeyTypeBlsBn254},
		{0x10, KeyTypeEcdsaSecp256k1},
		{0x20, KeyTypeBls12381},
		{0xFF, KeyTypeInvalid},
	}
	for _, c := range cases {
		if got := c.input.Type(); got != c.expected {
			t.Errorf("KeyTag.Type() = %v, want %v", got, c.expected)
		}
	}
}

func TestKeyTag_MarshalText(t *testing.T) {
	cases := []struct {
		input  KeyTag
		substr string
	}{
		{0x00, "BLS-BN254/0"},
		{0x0F, "BLS-BN254/15"},
		{0x10, "ECDSA-SECP256K1/0"},
		{0x1F, "ECDSA-SECP256K1/15"},
		{0x20, "BLS12-381/0"},
		{0xFF, "UNKNOWN/15"},
	}
	for _, c := range cases {
		b, err := c.input.MarshalText()
		if err != nil {
			t.Errorf("MarshalText() error: %v", err)
		}
		if !strings.Contains(string(b), c.substr) {
			t.Errorf("MarshalText() = %s, want substring %s", b, c.substr)
		}
	}
}

func TestKeyTag_String(t *testing.T) {
	cases := []struct {
		input  KeyTag
		substr string
	}{
		{0x00, "BLS-BN254/0"},
		{0x10, "ECDSA-SECP256K1/0"},
		{0x1F, "ECDSA-SECP256K1/15"},
		{0x20, "BLS12-381/0"},
		{0xFF, "UNKNOWN/15"},
	}
	for _, c := range cases {
		str := c.input.String()
		if !strings.Contains(str, c.substr) {
			t.Errorf("String() = %s, want substring %s", str, c.substr)
		}
	}
}

func TestKeyType_String(t *testing.T) {
	cases := []struct {
		input    KeyType
		expected string
	}{
		{KeyTypeBlsBn254, BLS_BN254_TYPE},
		{KeyTypeEcdsaSecp256k1, ECDSA_SECP256K1_TYPE},
		{KeyTypeBls12381, BLS12_381_TYPE},
		{KeyTypeInvalid, INVALID_TYPE},
	}
	for _, c := range cases {
		str, err := c.input.String()
		if err != nil {
			t.Errorf("KeyType.String() error: %v", err)
		}
		if str != c.expected {
			t.Errorf("KeyType.String() = %s, want %s", str, c.expected)
		}
	}
}

func TestKeyTypeFromString(t *testing.T) {
	cases := []struct {
		input    string
		expected KeyType
		wantErr  bool
	}{
		{BLS_BN254_TYPE, KeyTypeBlsBn254, false},
		{ECDSA_SECP256K1_TYPE, KeyTypeEcdsaSecp256k1, false},
		{BLS12_381_TYPE, KeyTypeBls12381, false},
		{"invalid", KeyTypeInvalid, false},
	}
	for _, c := range cases {
		kt, err := KeyTypeFromString(c.input)
		if (err != nil) != c.wantErr {
			t.Errorf("KeyTypeFromString(%s) error = %v, wantErr %v", c.input, err, c.wantErr)
		}
		if kt != c.expected && !c.wantErr {
			t.Errorf("KeyTypeFromString(%s) = %v, want %v", c.input, kt, c.expected)
		}
	}
}
