protoc -I=. --micro_out=. --go_out=. currencyrates.proto
protoc-go-inject-tag -input=currencyrates.pb.go -XXX_skip=bson,json,structure,validate
