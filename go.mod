module github.com/atomix/go-framework

go 1.12

require (
	github.com/atomix/api v0.0.0-20200129205515-343e74c131fa
	github.com/atomix/go-client v0.0.0-20200124004211-e5e19cd4730d
	github.com/atomix/go-local v0.0.0-20200124003802-357f6682b2f4
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	google.golang.org/grpc v1.27.0
)

replace github.com/atomix/api => ../api

replace github.com/atomix/go-client => ../go-client
