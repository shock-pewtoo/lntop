#!/bin/bash

mkdir -p built
docker build -t lntop-builder .
docker run --rm -it -v $PWD/built:/built lntop-builder
