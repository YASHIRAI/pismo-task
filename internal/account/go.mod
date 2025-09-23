module github.com/YASHIRAI/pismo-task/internal/account

go 1.21

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/YASHIRAI/pismo-task/internal/common v0.0.0-00010101000000-000000000000
	github.com/YASHIRAI/pismo-task/proto/account v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.8.4
)

replace github.com/YASHIRAI/pismo-task/internal/common => ../common

replace github.com/YASHIRAI/pismo-task/proto/account => ../../proto/account

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
