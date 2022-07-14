module github.com/opencord/voltha-northbound-bbf-adapter

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/Shopify/sarama v1.29.1
	github.com/golang/protobuf v1.5.2
	github.com/opencord/voltha-lib-go/v7 v7.2.0
	github.com/opencord/voltha-protos/v5 v5.2.3
	github.com/stretchr/testify v1.7.1
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
)
