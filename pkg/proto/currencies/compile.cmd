protoc -I=. --micro_out=. --go_out=. currencies.proto
protoc-go-inject-tag -input=currencies.pb.go -XXX_skip=bson,json,structure,validate

mockery -name=CurrencyratesService -recursive=true -output=../../mocks
