package shecomp

import (
	"errors"
	"strings"
	"testing"
)

func TestErrLargePlainText(t *testing.T) {
	// Test ErrLargePlainText.
	// To shorten the test time, override a private field: paddingReader. maxBitLength.

	// To check the boundary of the error, read twice.
	// The first time, check that no error occurs, and the second time, check that an error occurs.
	s := strings.Repeat("00", 2*blockSize)
	br := newPaddingReader(strings.NewReader(s))
	br.readBytes = maxBitLength/8 - blockSize

	o := make([]byte, blockSize)

	if err := br.block(o); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if err := br.block(o); !errors.Is(err, ErrLargePlainText) {
		t.Errorf("Expected ErrLargePlainText, got %v", err)
	}
}
