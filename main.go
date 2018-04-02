package main

import (
	"compress/zlib"
	"flag"
	"io"
	"log"
	"os"
	"time"

	csd "github.com/toravir/csd/libs"
)

func main() {
	inFile := flag.String("in", "<stdin>", "Input File (cbor Encoded)")
	outFile := flag.String("out", "<stdout>", "Output File to which decoded JSON will be written to (WILL overwrite if already present).")
	compressedIn := flag.Bool("compress", false, "Use if input stream is zlib compressed")

	flag.Parse()

	csd.DecodeTimeZone, _ = time.LoadLocation("America/Los_Angeles")
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	if *inFile != "<stdin>" {
		f, err := os.Open(*inFile)
		if err != nil {
			log.Fatal(err)
		}
		in = f
		defer func() {
			f.Close()
		}()
	}
	if *compressedIn {
		zin, err := zlib.NewReader(in)
		if err != nil {
			log.Fatal(err)
		}
		in = zin
		defer func() {
			zin.Close()
		}()
	}
	if *outFile != "<stdout>" {
		f, err := os.OpenFile(*outFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		out = f
		defer func() {
			f.Close()
		}()
	}
	csd.Cbor2JsonManyObjects(in, out)
}
