module cosmossdk.io/x/tx

go 1.22

toolchain go1.22.2

require (
	cosmossdk.io/api v0.7.2
	cosmossdk.io/core v0.11.0
	cosmossdk.io/errors v1.0.0-beta.7
	cosmossdk.io/math v1.2.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.3
	github.com/ethereum/go-ethereum v1.13.10
	github.com/google/go-cmp v0.6.0
	github.com/iancoleman/strcase v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.34.1
	gotest.tools/v3 v3.5.0
	pgregory.net/rapid v1.1.0
)

require (
	github.com/cosmos/gogoproto v1.4.11 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// HV2 related packages
replace (
	cosmossdk.io/api => github.com/0xPolygon/cosmos-sdk/api v0.7.2
	cosmossdk.io/core => github.com/0xPolygon/cosmos-sdk/core v0.11.0
	cosmossdk.io/errors => github.com/0xPolygon/cosmos-sdk/errors v1.0.0-beta.7
	cosmossdk.io/math => github.com/0xPolygon/cosmos-sdk/math v1.2.0

	github.com/ethereum/go-ethereum => github.com/maticnetwork/bor v1.4.0
)
