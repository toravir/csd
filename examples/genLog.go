package main

import (
	"fmt"
	"io"
//	"net"
	"os"
	"strconv"
	"time"
    "log"

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

	// https://www.cisco.com/c/en/us/td/docs/switches/datacenter/aci/apic/sw/1-x/faults/guide/b_APIC_Faults_Errors/b_IFC_Faults_Errors_chapter_011.html#concept_FC4DA4AA93FB442E9D5940E5F52A4FEF
	//
	// May 22 15:49:54 192.168.10.102 <1027> May 22 22:49:54 spine1
	// %LOG_LOCAL0-3-SYSTEM_MSG
	//[F41650][raised][threshold-crossed][major][sys/ch/scslot-6/sc/sensor-1/fault-F41650]
	//TCA: eqptTemp5min normalizedLast value 84 raised above threshold 80

	zerolog.TimestampFunc = func() time.Time { return time.Now().Round(time.Second) }
	log := zerolog.New(f).With().
		//IPAddr("IP", net.IP{192, 168, 10, 102}).
		Timestamp().
		Logger()
	for i := 0; i < count; i++ {
		log.Error().
			Int("Fault", 41650+i).Msg("TCA: eqptTemp5min normalizedLast value 84 raised above threshold 80")
		//time.Sleep(time.Second)
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
