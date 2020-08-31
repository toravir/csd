package main

import (
	"compress/zlib"
	"flag"
	"io"
	//	"net"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func writeLog(fname string, count int, useCompress bool) {
	opFile := os.Stdout
	if fname != "<stdout>" {
		fil, _ := os.Create(fname)
		opFile = fil
		defer func() {
			if err := fil.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	var f io.WriteCloser = opFile
	if useCompress {
		f = zlib.NewWriter(f)
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()

	}

	zerolog.TimestampFunc = func() time.Time { return time.Now().Round(time.Second) }
	log := zerolog.New(f).With().
		//IPAddr("IP", net.IP{192, 168, 10, 102}).
		Timestamp().
		Logger()
	for i := 0; i < count; i++ {
		time.Sleep(time.Duration(i%5) * time.Second)
		log.Error().
			Int("Fault", 41650+i).Msg("TCA:")
	}
}

func main() {
	outFile := flag.String("out", "<stdout>", "Output File to which logs will be written to (WILL overwrite if already present).")
	numLogs := flag.Int("num", 10, "Number of log messages to generate.")
	doCompress := flag.Bool("compress", false, "Enable inline compressed writer")

	flag.Parse()

	writeLog(*outFile, *numLogs, *doCompress)
}
