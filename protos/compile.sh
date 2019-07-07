#!/bin/bash

proto_dir="$(dirname "$BASH_SOURCE")"
cd "$proto_dir"
protoc *.proto --go_out=plugins=grpc:.