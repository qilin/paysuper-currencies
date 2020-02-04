module github.com/paysuper/paysuper-currencies

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/ProtocolONE/rabbitmq v0.0.0-20190129162844-9f24367e139c
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gogo/protobuf v1.2.1
	github.com/golang-migrate/migrate/v4 v4.3.1
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins/client/selector/static v0.0.0-20200119172437-4fe21aa238fd
	github.com/micro/go-plugins/wrapper/monitoring/prometheus v0.0.0-20200119172437-4fe21aa238fd
	github.com/paysuper/paysuper-database-mongo v0.1.1
	github.com/paysuper/paysuper-proto/go/currenciespb v0.0.0-20200203130641-45056764a1d7
	github.com/paysuper/paysuper-tools v0.0.0-20200116214558-6afcd9131e1c
	github.com/prometheus/client_golang v1.3.0
	github.com/satori/go.uuid v1.2.0
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.4.0
	github.com/thetruetrade/gotrade v0.0.0-20140906064133-08b7c41e93d9
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/go-playground/validator.v9 v9.30.0
)

replace github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d

go 1.13
