#!/bin/bash
cd "$(dirname "$0")"
rm -f hysteresis
go build -o hysteresis ./cmd/hysteresis && ./hysteresis
