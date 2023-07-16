package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tenkoh/go-shecomp"
)

func TestCompress(t *testing.T) {
	var b bytes.Buffer
	s := "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51"
	want := "c7277a0dc1fb853b5f4d9cbd26be40c6"

	if err := compress(&b, strings.NewReader(s), shecomp.Compress); err != nil {
		t.Error(err)
		return
	}

	if b.String() != want {
		t.Errorf("got %s, want %s", b.String(), want)
	}
}
