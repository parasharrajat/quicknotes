#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

rm -f resources.go
wc -l s/*.html sass/*.sass sass/*.scss
echo && wc -l js/*.ts*
echo && wc -l *.go
