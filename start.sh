#!/bin/bash -e

GOROOT=$(go env GOROOT)
GOROOT_SRC=$GOROOT/src

mkdir -p $GOROOT_SRC/toolkit/pkg
cp -r pkg/trace $GOROOT_SRC/toolkit/pkg