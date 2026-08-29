module github.com/complexus-tech/projects-api/examples/external-integration

go 1.24.0

require (
	github.com/complexus-tech/fortyone-go v0.0.0
	github.com/google/uuid v1.6.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/oapi-codegen/runtime v1.6.0 // indirect
)

replace github.com/complexus-tech/fortyone-go => ../../sdk/go
