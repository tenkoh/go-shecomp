# go-shecomp
This repository provides an implementation of the Compression function defined by AUTOSAR - Secure Hardware Extensions (SHE).


Reference: [AUTOSAR - Specification of Secure Hardware Extensions](https://www.autosar.org/fileadmin/standards/R22-11/FO/AUTOSAR_TR_SecureHardwareExtensions.pdf)

This repository includes a Go's package and a command-line tool. You can utilize the package to implement your own application, or you can use the command-line tool to experiment with the compression function.

## Install
To install, visit the [release page](https://github.com/tenkoh/go-shecomp/releases) and download the latest binary.

If you are already familiar to Golang, you can install it by `go install` command.

```bash
go install github.com/tenkoh/go-shecomp/cmd/shecomp@latest
```

## Usage

### Package
Simply use the `shecomp.Compress` function.

```go
package main

import (
    "fmt"
    "strings"

    "github.com/tenkoh/go-shecomp"
)

func main() {
    // the input must be hexadecimal encoded.
    r := strings.NewReader("0123456789abcdef")
    compressed, err := shecomp.Compress(r)
    if err != nil {
        panic(err)
    }

    // the output above is hexadecimal encoded.
    fmt.Println(string(compressed))
}
```

If you only require padding, you can use the `shecomp.Padding` function.


### Command-line tool
The command-line tool accepts hexadecimal encoded data from the terminal, either as an argument or through stdin, and outputs the hexadecimal encoded data to stdout.

The basic usage is as follows:
```bash
shecomp {hexadecimal encoded data}
```

For example, to compress `0123456789abcdef`, use the following command:
```bash
shecomp 0123456789abcdef
# get 8cc511383f521cb60a9b8b0358e7e17d
```

You can also use pipes:
```bash
cat input.txt | shecomp

# or

shecomp < input.txt
```

To specify the input file, use the `--input` flag:
```bash
shecomp --input input.txt
```

If you only want padding, use the `--padding` flag:
```bash
shecomp --padding {hexadecimal encoded data}
```

For more information, refer to the help section:
```bash
shecomp --help
```

## Contribution
Contributions are welcome. Please submit pull requests or issues.

## License
MIT

## Author
tenkoh
