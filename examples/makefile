all: genLogJson genLogBin

genLogJson: genLog.go
	@echo "FYI: These examples use APIs from github.com/rs/zerolog library to generate log msgs in CBOR/JSON format"
	go build -o genLogJson genLog.go 

genLogBin: genLog.go
	@echo "FYI: These examples use APIs from github.com/rs/zerolog library to generate log msgs in CBOR/JSON format"
	go build -tags binary_log -o genLogBin genLog.go 

clean:
	rm -f genLogJson genLogBin
