#!/bin/bash
set -u -e -o pipefail

cd cmd/blogtojson
go tool vet -printfuncs=LogInfof,LogErrorf,LogVerbosef .
go build -o blogtojson
cd ../..
./cmd/blogtojson/blogtojson -dir=../blog/blog_posts -out=blog.json || true
rm ./cmd/blogtojson/blogtojson
