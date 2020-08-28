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

var FileFollowPollInterval = 5 * time.Second

type followReader struct {
    f *os.File
    follow bool
}

func NewFollowReader(fname string, follow bool) (*followReader, error) {
    var err error
    f := &followReader{}
    f.f, err = os.Open(fname)
    if err != nil {
        return nil, err
    }
    f.follow = follow
    return f, nil
}

func (f *followReader) Read(p []byte) (int, error) {
    for {
        n, err := f.f.Read(p)
        if err == nil {
            return n, err
        }
        if f.follow && err == io.EOF {
            time.Sleep(FileFollowPollInterval)
        } else {
            return n, err
        }
    }
}

func (f *followReader) Close() {
    f.f.Close()
}

func main() {
	inFile := flag.String("in", "<stdin>", "Input File (cbor Encoded)")
	outFile := flag.String("out", "<stdout>", "Output File to which decoded JSON will be written to (WILL overwrite if already present).")
	compressedIn := flag.Bool("compress", false, "Use if input stream is zlib compressed")
	follow := flag.Bool("follow", false, "tail the file (default for stdin)")

	flag.Parse()

	csd.DecodeTimeZone, _ = time.LoadLocation("America/Los_Angeles")
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	if *inFile != "<stdin>" {
		f, err := NewFollowReader(*inFile, *follow)
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
