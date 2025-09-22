module github.com/YASHIRAI/pismo-task/services/transaction-manager

go 1.21

require (
	github.com/YASHIRAI/pismo-task/api/proto/transaction v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.64.0
)

replace github.com/YASHIRAI/pismo-task/api/proto/transaction => ../../api/proto/transaction

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
