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

// ErrNeedPadding is returned when the input text is not multiple of block size.
var ErrNeedPadding = errors.New("shecomp: the input text must be multiple of block size")

type blockReader interface {
	block(dst []byte) error
}

func hexDecode(dst []byte, src io.Reader) (int, error) {
	h := make([]byte, hex.EncodedLen(len(dst)))
	n, err := src.Read(h)
	if err != nil {
		return 0, err
	}
	h = h[:n]
	return hex.Decode(dst, h)
}

type noPaddingReader struct {
	r io.Reader
}

type paddingReader struct {
	r         io.Reader
	b         []byte
	readBytes uint64
	pad       []byte
	eof       bool
}

func (r *noPaddingReader) block(dst []byte) error {
	n, err := hexDecode(dst, r.r)
	if err != nil {
		return err
	}
	if n < blockSize {
		return ErrNeedPadding
	}
	return nil
}

func newPaddingReader(r io.Reader) *paddingReader {
	return &paddingReader{
		r: r,
		b: make([]byte, blockSize),
	}
}

func (r *paddingReader) block(dst []byte) error {
	if r.eof {
		return io.EOF
	}
	// read into r.b and copy from r.b to dst
	n, err := hexDecode(r.b, r.r)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	r.readBytes += uint64(n)

	if r.readBytes*8 > maxBitLength {
		return ErrLargePlainText
	}

	if n == blockSize {
		copy(dst, r.b)
		return nil
	}

	// calculate padding bytes
	r.eof = true
	r.b = r.b[:n]
	r.pad = padding(r.b, r.readBytes)

	copy(dst, append(r.b, r.pad...))
	return nil
}

// compress compresses the input data using AES Miyaguchi-Preenel mode.
// This function returns both the compressed data and the padding bytes.
// The input data must be hexadecimal encoded.
func compress(br blockReader) ([]byte, error) {
	src := make([]byte, blockSize)
	out := make([]byte, blockSize)

	for {
		if err := br.block(src); err != nil {
			if errors.Is(err, io.EOF) {
				return out, nil
			}
			return nil, fmt.Errorf("could not read from reader: %w", err)
		}

		o, err := encrypt(src, out)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt: %w", err)
		}
		out = o
	}
}

// Compress compresses the input data using AES Miyaguchi-Preenel mode.
// The output is encoded in hexadecimal.
// The input data must be hexadecimal encoded.
// If the length of the input text is greater than 1<<40 - 1 in bit, it returns ErrLargePlainText.
func Compress(r io.Reader) ([]byte, error) {
	br := newPaddingReader(r)
	c, err := compress(br)
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
	br := newPaddingReader(r)
	out := make([]byte, blockSize)
	for {
		if err := br.block(out); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("error during padding: %w", err)
		}
	}
	h := make([]byte, hex.EncodedLen(len(br.pad)))
	hex.Encode(h, br.pad)
	return h, nil
}

// CompressWithoutPadding compresses the input data using AES Miyaguchi-Preenel mode.
// CompressWithoutPadding is almost same as Compress function, but does not add padding to the end of the input data.
// The input data must have appropriate padding according to the SHE protocol.
func CompressWithoutPadding(r io.Reader) ([]byte, error) {
	br := noPaddingReader{r}
	c, err := compress(&br)
	if err != nil {
		return nil, err
	}
	h := make([]byte, hex.EncodedLen(len(c)))
	hex.Encode(h, c)
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
