package shecomp_test

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/tenkoh/go-shecomp"
)

func TestDecodeHexStream(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *shecomp.PlainText
		wantErr bool
	}{
		{"empty input", "", &shecomp.PlainText{Blocks: [][]byte{}, Remain: []byte{}}, false},
		{"invalid input, odd number digits hex", "8", nil, true},
		{
			"valid input, only remain",
			"88",
			&shecomp.PlainText{
				Blocks: [][]byte{},
				Remain: []byte{0x88},
			},
			false,
		},
		{
			"valid input, only blocks",
			strings.Repeat("88", 16),
			&shecomp.PlainText{
				Blocks: [][]byte{bytes.Repeat([]byte{0x88}, 16)},
				Remain: []byte{},
			},
			false,
		},
		{
			"valid input",
			strings.Repeat("88", 17),
			&shecomp.PlainText{
				Blocks: [][]byte{bytes.Repeat([]byte{0x88}, 16)},
				Remain: []byte{0x88},
			},
			false,
		},
		{
			"invalid input, with newline",
			"88\r\n",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := shecomp.DecodeHexStream(strings.NewReader(tt.input))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexptected error: %v", err)
				}
				return
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("DecodeHexStream() = %v, want %v", got, tt.want)
			}
		})
	}

	// TODO: add test for long input
}

func TestPadding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"empty input", "", hexMustDecode(t, "8"+strings.Repeat("0", 31))},
		{
			"input's length is multiple of 128bit",
			strings.Repeat("88",
				32),
			hexMustDecode(t, "8"+strings.Repeat("0", 28)+"100"),
		},
		{
			"the zero part of padding is the shortest",
			strings.Repeat("88", 26),
			hexMustDecode(t, "8"+strings.Repeat("0", 9)+"D0"),
		},
		{
			"example on SHE specification",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51",
			hexMustDecode(t, "8"+strings.Repeat("0", 28)+"100"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pt, err := shecomp.DecodeHexStream(strings.NewReader(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			got := shecomp.Padding(pt)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Padding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func hexMustDecode(t *testing.T, s string) []byte {
	t.Helper()
	h, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			"example described in SHE specification",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51",
			hexMustDecode(t, "c7277a0dc1fb853b5f4d9cbd26be40c6"),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pt, err := shecomp.DecodeHexStream(strings.NewReader(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			got, err := shecomp.Compress(pt)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Compress() = %v, want %v", got, tt.want)
			}
		})
	}
}
