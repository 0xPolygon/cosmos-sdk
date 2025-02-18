module cosmossdk.io/x/tx

go 1.23.2

require (
	cosmossdk.io/api v0.7.4
	cosmossdk.io/core v0.11.0
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/math v1.4.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/ethereum/go-ethereum v1.15.0
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// NOTE: we do not want to replace to the development version of cosmossdk.io/api yet
// Until https://github.com/cosmos/cosmos-sdk/issues/19228 is resolved
// We are tagging x/tx v0.14+ from main and v0.13 from release/v0.50.x and must keep using released versions of x/tx dependencies

// HV2 related packages
replace (
	cosmossdk.io/api => github.com/0xPolygon/cosmos-sdk/api v0.7.4
	cosmossdk.io/core => github.com/0xPolygon/cosmos-sdk/core v0.11.3-0.20241126102051-89dc71d02611
	cosmossdk.io/errors => github.com/0xPolygon/cosmos-sdk/errors v1.0.0-beta.7.0.20241126102051-89dc71d02611
	cosmossdk.io/math => github.com/0xPolygon/cosmos-sdk/math v1.4.0
	github.com/ethereum/go-ethereum => github.com/maticnetwork/bor v1.5.5
)
