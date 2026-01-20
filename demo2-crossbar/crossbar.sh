#!/bin/bash
cd "$(dirname "$0")"
go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui
