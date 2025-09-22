module github.com/YASHIRAI/pismo-task/cmd/account-mgr

go 1.21

require (
	github.com/YASHIRAI/pismo-task/internal/account v0.0.0-00010101000000-000000000000
	github.com/YASHIRAI/pismo-task/internal/common v0.0.0-00010101000000-000000000000
	github.com/YASHIRAI/pismo-task/proto/account v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
)

replace github.com/YASHIRAI/pismo-task/internal/account => ../../internal/account

replace github.com/YASHIRAI/pismo-task/internal/common => ../../internal/common

replace github.com/YASHIRAI/pismo-task/proto/account => ../../proto/account

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)
