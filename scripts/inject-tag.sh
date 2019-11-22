#!/usr/bin/env sh

if [ -n "$1" ] && [ ${0:0:4} = "/bin" ]; then
  ROOT_DIR=$1/..
else
  ROOT_DIR="$( cd "$( dirname "$0" )" && pwd )/.."
fi

. $ROOT_DIR/scripts/common.sh

protoc-go-inject-tag -input=${PROTO_GEN_PATH}/currencies/currencies.pb.go -XXX_skip=bson,json,structure,validate