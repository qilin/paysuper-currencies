module github.com/paysuper/paysuper-currencies

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/InVisionApp/go-logger v1.0.1 // indirect
	github.com/ProtocolONE/rabbitmq v0.0.0-20190129162844-9f24367e139c
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/favadi/protoc-go-inject-tag v0.0.0-20181008023834-c2c1884c833d // indirect
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gogo/protobuf v1.2.1
	github.com/golang-migrate/migrate/v4 v4.3.1
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/paysuper/paysuper-database-mongo v0.1.1
	github.com/paysuper/paysuper-proto/go/currenciespb v0.0.0-20200119002105-6af7d2096854
	github.com/paysuper/paysuper-tools v0.0.0-20200116214558-6afcd9131e1c
	github.com/prometheus/client_golang v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.4.0
	github.com/thetruetrade/gotrade v0.0.0-20140906064133-08b7c41e93d9
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
)

go 1.13
