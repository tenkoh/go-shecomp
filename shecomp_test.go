package shecomp_test

import (
	"bytes"
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

}

func TestCompress(t *testing.T) {

}
