package shecomp_test

import (
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
			[]byte("8" + strings.Repeat("0", 31)),
			false,
		},
		{
			"input's length is multiple of 128bit",
			strings.Repeat("88", 32),
			[]byte("8" + strings.Repeat("0", 28) + "100"),
			false,
		},
		{
			"the zero part of padding is the shortest",
			strings.Repeat("88", 26),
			[]byte("8" + strings.Repeat("0", 9) + "d0"),
			false,
		},
		{
			"example on SHE specification",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51",
			[]byte("8" + strings.Repeat("0", 28) + "100"),
			false,
		},
		{
			"invalid hexadecimal input",
			"012",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			[]byte("c7277a0dc1fb853b5f4d9cbd26be40c6"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestCompressWithoutPadding(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			"example described in SHE specification 4.13.2.10",
			"000102030405060708090a0b0c0d0e0f010153484500800000000000000000b0",
			[]byte("118a46447a770d87828a69c222e2d17e"),
			false,
		},
		{
			"the input is not multiple of block size",
			"0000000000000000000000000000000",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shecomp.CompressWithoutPadding(strings.NewReader(tt.input))
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
