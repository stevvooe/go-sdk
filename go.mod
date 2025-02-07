module github.com/statsig-io/go-sdk

go 1.16

replace github.com/statsig-io/go-sdk/internal => ./internal/

replace github.com/statsig-io/go-sdk/types => ./types/

require (
	github.com/statsig-io/ip3country-go v0.2.0
	github.com/ua-parser/uap-go v0.0.0-20210121150957-347a3497cc39
)
