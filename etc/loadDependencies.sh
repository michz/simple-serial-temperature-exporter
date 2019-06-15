#!/usr/bin/env bash

PROJECT_DIR="$( cd "$( dirname "$( dirname "${BASH_SOURCE[0]}")" )" && pwd )"

cd ${PROJECT_DIR}

go get github.com/jacobsa/go-serial/serial
#go get github.com/argandas/serial
