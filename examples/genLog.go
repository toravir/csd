package main

import (
	"fmt"
	"io"
	//	"net"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func writeLog(fname string, count int) {
	opFile := os.Stdout
	if fname != "-" {
		fil, _ := os.Create(fname)
		opFile = fil
	}
	var f io.WriteCloser = opFile
	defer func() {
		if err := opFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	zerolog.TimestampFunc = func() time.Time { return time.Now().Round(time.Second) }
	log := zerolog.New(f).With().
		//IPAddr("IP", net.IP{192, 168, 10, 102}).
		Timestamp().
		Logger()
	for i := 0; i < count; i++ {
		log.Error().
			Int("Fault", 41650+i).Msg("TCA:")
	}
}

func printUsage() {
	fmt.Println("Usage: genLog <outputFile> [<count>]")
	fmt.Println("default value for count is 100")
}

func main() {
	writeCount := 100

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(-1)
	} else {
		for i := 2; i < len(os.Args); i++ {
			count, ok := strconv.Atoi(os.Args[2])
			if ok != nil {
				count = 100
			}
			writeCount = count
		}
	}
	outputFile := os.Args[1]

	writeLog(outputFile, writeCount)
}
