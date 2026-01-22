.PHONY: build test run cli clean

build:
	go build ./...

test:
	go test ./... -v

run:
	go run ./cmd/eda-gui

cli:
	go run ./cmd/eda-cli -input data/sample_weights_8x8.json -rows 8 -cols 8

clean:
	go clean
	rm -f crossbar_mapping.json cell_assignments.csv crossbar.sp
