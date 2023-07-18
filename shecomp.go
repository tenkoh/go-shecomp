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

type hexVerifiedDecoder struct {
	r         io.Reader
	blockSize int
}

func newHexVerifiedDecoder(r io.Reader, blockSize int) *hexVerifiedDecoder {
	return &hexVerifiedDecoder{r, blockSize}
}

// *hexVerifiedDecoder.Read is different from hex.Decoder.Read
// at the point of returning a decode error immediately. ex)ErrLength.
func (d *hexVerifiedDecoder) read(dst []byte) (int, error) {
	h := make([]byte, hex.EncodedLen(len(dst)))
	n, err := d.r.Read(h)
	if err != nil {
		return 0, err
	}
	h = h[:n]
	return hex.Decode(dst, h)
}

// compress compresses the input data using AES Miyaguchi-Preenel mode.
// This function returns both the compressed data and the padding bytes.
// The input data must be hexadecimal encoded.
func compress(r io.Reader) (compressed, pad []byte, _ error) {
	d := newHexVerifiedDecoder(r, blockSize)
	src := make([]byte, blockSize)
	out := make([]byte, blockSize)

	var srcByteLen uint64
	for {
		n, err := d.read(src)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, nil, fmt.Errorf("shecomp: failed to read input stream: %w", err)
			}
		}
		srcByteLen += uint64(n)
		src = src[:n]
		if n < blockSize || errors.Is(err, io.EOF) {
			break
		}
		o, err := encrypt(src, out)
		if err != nil {
			return nil, nil, fmt.Errorf("shecomp: failed to encrypt: %w", err)
		}
		out = o
	}
	if srcByteLen*8 > maxBitLength {
		return nil, nil, ErrLargePlainText
	}
	pad = padding(src, srcByteLen)
	out, err := encrypt(append(src, pad...), out)
	if err != nil {
		return nil, nil, fmt.Errorf("shecomp: failed to encrypt the last block with padding: %w", err)
	}
	return out, pad, nil
}

// Compress compresses the input data using AES Miyaguchi-Preenel mode.
// The output is encoded in hexadecimal.
// The input data must be hexadecimal encoded.
// If the length of the input text is greater than 1<<40 - 1 in bit, it returns ErrLargePlainText.
func Compress(r io.Reader) ([]byte, error) {
	c, _, err := compress(r)
	if err != nil {
		return nil, err
	}
	h := make([]byte, hex.EncodedLen(len(c)))
	hex.Encode(h, c)
	return h, nil
}

// Padding calculate the padding bytes.
// The output is encoded in hexadecimal.
// This function does not modify the input, just returns the padding bytes.
// The input data must be hexadecimal encoded.
// If the length of the input text is greater than 1<<40 - 1 in bit, it returns ErrLargePlainText.
func Padding(r io.Reader) ([]byte, error) {
	_, pad, err := compress(r)
	if err != nil {
		return nil, err
	}
	h := make([]byte, hex.EncodedLen(len(pad)))
	hex.Encode(h, pad)
	return h, nil
}

func encrypt(src, previous []byte) ([]byte, error) {
	if len(src) != blockSize || len(previous) != blockSize {
		return nil, fmt.Errorf("shecomp: failed to encrypt. the length of each input must be same as blockSize=%d, but len(src) = %d, len(previous) = %d", blockSize, len(src), len(previous))
	}
	cipher, err := aes.NewCipher(previous)
	if err != nil {
		return nil, fmt.Errorf("shecomp: failed to create cipher: %w", err)
	}
	encrypted := make([]byte, blockSize)
	cipher.Encrypt(encrypted, src)
	encrypted, _ = xor(encrypted, src)
	out, _ := xor(previous, encrypted)
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

func padding(b []byte, messageByteLen uint64) []byte {
	padMinBitLen := 8*len(b) + 1 + 40
	padByteLen := (padMinBitLen/128+1)*128/8 - len(b)
	pad := make([]byte, padByteLen)

	// the last 40 bits of the padding shows the length of the message in bits
	for i := 0; i < 5; i++ {
		pos := len(pad) - 1 - i
		d := uint8(0xff & ((messageByteLen * 8) >> (8 * i)))
		pad[pos] = d
	}
	// the first bit of the padding must be 1
	pad[0] = 0x80 | pad[0]
	return pad
}
