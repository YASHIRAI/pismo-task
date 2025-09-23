module github.com/YASHIRAI/pismo-task/internal/account

go 1.24.0

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
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250922171735-9219d122eba9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
