#!/bin/bash
cd "$(dirname "$0")/demo2-crossbar"
go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui
