#!/bin/sh
set -e
cd "$(dirname "$(readlink -f "$0")")"
go run go.elastic.co/fastjson/cmd/generate-fastjson -f -o marshal_fastjson.go .
exec go run github.com/elastic/go-licenser marshal_fastjson.go
