// Package shecomp provides packages and cli
// which implements Compression function defined by
// AUTOSAR Secure Hardware Extensions (SHE).
package shecomp

import (
	"crypto/aes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	blockSize    = 16 // 128 bits
	maxBitLength = 1<<40 - 1
)

// ErrLargePlainText is returned when the length of the input text is greater than 1<<40 - 1.
var ErrLargePlainText = errors.New("shecomp: the length of the input text is too large")

// PlainText is a struct which split a huge message into blocks with constant size.
// The remained has zero or more bytes less than the block size.
type PlainText struct {
	Blocks [][]byte
	Remain []byte // remaining piece after splitting whole message into some blocks
}

// bitLength returns the length of the PlainText in bits.
// The returned value's type is uint64 which is enough to store the length of the PlainText
// which is defined by the AUTOSAR SHE protocol.
func (pt *PlainText) bitLength() uint64 {
	var sum uint64
	sum += uint64(16 * len(pt.Blocks))
	sum += uint64(len(pt.Remain))
	sum *= 8
	return sum
}

func newPlainText() *PlainText {
	return &PlainText{
		Blocks: make([][]byte, 0),
		Remain: make([]byte, 0),
	}
}

// DecodeHexStream reads and splits a hex stream from r, deocode each blocks into []byte
// and then returns a PlainText.
func DecodeHexStream(r io.Reader) (*PlainText, error) {
	pt := newPlainText()
	for {
		b := make([]byte, blockSize*2)
		n, err := r.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		b = b[:n]
		h := make([]byte, n/2)

		if _, err := hex.Decode(h, b); err != nil {
			return nil, err
		}
		if n == blockSize*2 {
			pt.Blocks = append(pt.Blocks, h)
		} else {
			pt.Remain = h
		}
	}
	if pt.bitLength() > maxBitLength {
		return nil, ErrLargePlainText
	}
	return pt, nil
}

// Padding calculate padding bytes. This function does not modify the original PlainText.
func Padding(pt *PlainText) []byte {
	r := pt.Remain
	remainBits := 8 * len(r)
	padMinBits := remainBits + 1 + 40 // defined on the SHE protocol
	padBytes := (padMinBits/128+1)*128/8 - len(r)
	pad := make([]byte, padBytes)

	// the lower 40bits = 5bytes shows the plain text's bit length in uint expression
	ptBitLength := pt.bitLength()
	for i := 0; i < 5; i++ {
		pos := len(pad) - 1 - i
		d := uint8(0xff & (ptBitLength >> (8 * i)))
		pad[pos] = d
	}

	// the head bit is 1.
	pad[0] = 0x80 | pad[0]

	return pad
}

// Compress calculate a secure hash value of a PlainText using AES Miyaguchi-Preneel mode.
func Compress(pt *PlainText) ([]byte, error) {
	pad := Padding(pt)
	padded := append(pt.Remain, pad...)
	src := append(pt.Blocks, padded)
	out := make([]byte, blockSize)
	encrypted := make([]byte, blockSize)
	for i := 0; i < len(src); i++ {
		cipher, err := aes.NewCipher(out)
		if err != nil {
			return nil, fmt.Errorf("shecomp: failed to create cipher of block = %d: %w", i, err)
		}
		cipher.Encrypt(encrypted, src[i])
		encrypted, _ = xor(encrypted, src[i])
		out, _ = xor(out, encrypted)
	}
	return out, nil
}

func xor(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("shecomp: failed to xor: len(a) = %d, len(b) = %d", len(a), len(b))
	}
	r := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}
	return r, nil
}
