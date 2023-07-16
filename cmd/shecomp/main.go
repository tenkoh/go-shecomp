// cmd/shecomp is a command line tool to compress the input data using Compress function in AUTOSAR Secure Hardware Extension (SHE).
package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/tenkoh/go-shecomp"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	version = "v0.0.1"
)

func compress(w io.Writer, r io.Reader, fn func(r io.Reader) ([]byte, error)) error {
	o, err := fn(r)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	enc := hex.NewEncoder(&b)
	n, err := enc.Write(o)
	if err != nil {
		return err
	}
	if n != len(o) {
		return errors.New("failed to hexadecimal encode")
	}

	fmt.Fprint(w, b.String())
	return nil
}

func run(c *cli.Context) error {
	// swtich the input source
	// if both the input file and the hexadecimal encoded string are specified, return error.
	if c.String("input") != "" && len(c.Args().Slice()) > 0 {
		return errors.New("both the input file and the hexadecimal encoded string are specified")
	}

	var r io.Reader
	r = os.Stdin
	if c.String("input") != "" {
		f, err := os.Open(c.String("input"))
		if err != nil {
			return fmt.Errorf("failed to open the input file %s: %w", c.String("input"), err)
		}
		defer f.Close()
		r = f
	}
	if len(c.Args().Slice()) > 0 {
		r = strings.NewReader(c.Args().Slice()[0])
	}

	// switch the output mode
	var fn func(r io.Reader) ([]byte, error)
	if c.Bool("padding") {
		fn = shecomp.Padding
	} else {
		fn = shecomp.Compress
	}

	if err := compress(os.Stdout, r, fn); err != nil {
		return err
	}

	// if the output is a terminal, add a new line
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println()
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:    "shecomp",
		Usage:   "shecomp [options] [hexadecimal encoded string]",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "specify the input file",
			},
			&cli.BoolFlag{
				Name:    "padding",
				Aliases: []string{"p"},
				Usage:   "switch output mode to return only padding",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
