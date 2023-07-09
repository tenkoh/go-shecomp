package shecomp_test

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/tenkoh/go-shecomp"
)

func TestPadding(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []byte
		wantError bool
	}{
		{
			"empty input",
			"",
			hexMustDecode(t, "8"+strings.Repeat("0", 31)),
			false,
		},
		{
			"input's length is multiple of 128bit",
			strings.Repeat("88", 32),
			hexMustDecode(t, "8"+strings.Repeat("0", 28)+"100"),
			false,
		},
		{
			"the zero part of padding is the shortest",
			strings.Repeat("88", 26),
			hexMustDecode(t, "8"+strings.Repeat("0", 9)+"D0"),
			false,
		},
		{
			"example on SHE specification",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51",
			hexMustDecode(t, "8"+strings.Repeat("0", 28)+"100"),
			false,
		},
		{
			"invalid hexdecimal input",
			"012",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got, err := shecomp.Padding(strings.NewReader(tt.input))
			if err != nil {
				if !tt.wantError {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
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
			got, err := shecomp.Compress(strings.NewReader(tt.input))
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
