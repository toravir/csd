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
	follow := flag.Bool("follow", false, "tail the file (default for stdin)")

	flag.Parse()

	csd.DecodeTimeZone, _ = time.LoadLocation("America/Los_Angeles")
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	ch := make(chan struct{})
	if *inFile != "<stdin>" {
		f, err := csd.NewFollowReader(*inFile, *follow, ch)
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
