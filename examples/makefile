all: genLogJson genLogBin

genLogJson: genLog.go
	go build -o genLogJson genLog.go 

genLogBin: genLog.go
	go build -tags binary_log -o genLogBin genLog.go 

clean:
	rm -f genLogJson genLogBin
