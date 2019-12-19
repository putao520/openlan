package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"strings"
	"github.com/pierrec/lz4"
)

func ExampleCompressBlock() {
	s := "hello world"
	data := []byte(strings.Repeat(s, 2))
	buf := make([]byte, len(data))
	ht := make([]int, 64<<10) // buffer for the compression table

	n, err := lz4.CompressBlock(data, buf, ht)
	if err != nil {
		fmt.Println(err)
	}
	if n >= len(data) {
		fmt.Printf("`%s` is not compressible", s)
	}

	fmt.Printf("<%d>\n", n)

	buf = buf[:n] // compressed data

	// Allocated a very large buffer for decompression.
	out := make([]byte, 10*len(data))
	n, err = lz4.UncompressBlock(buf, out)
	if err != nil {
		fmt.Println(err)
	}
	out = out[:n] // uncompressed data

	fmt.Println(string(out))

	// Output:
	// hello world
}

func main() {
	s := "hello world"
	data := []byte(strings.Repeat(s, 2))

	buf := libol.Lz4Compress(data)
	fmt.Printf("%s\n", buf)
	n := libol.Lz4Uncompress(buf, 10*len(data))

	fmt.Printf("%s\n", s)
	fmt.Printf("%s,%d\n", string(n), len(n))

	ExampleCompressBlock()
}
