module cosmossdk.io/store

go 1.22

toolchain go1.22.2

require (
	cosmossdk.io/errors v1.0.0
	cosmossdk.io/log v1.2.1
	cosmossdk.io/math v1.2.0
	github.com/cometbft/cometbft v0.38.0
	github.com/cosmos/cosmos-db v1.0.0
	github.com/cosmos/gogoproto v1.4.11
	github.com/cosmos/iavl v1.0.0
	github.com/cosmos/ics23/go v0.10.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-metrics v0.5.1
	github.com/hashicorp/go-plugin v1.5.2
	github.com/hashicorp/golang-lru v1.0.2
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8
	google.golang.org/grpc v1.64.1
	google.golang.org/protobuf v1.34.1
	gotest.tools/v3 v3.5.1
)

require (
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/emicklei/dot v1.4.2 // indirect
	github.com/ethereum/go-ethereum v1.13.4 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/getsentry/sentry-go v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/jhump/protoreflect v1.15.3 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.7.16 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220708102147-0a8a51822cae // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/petermattis/goid v0.0.0-20221215004737-a150e88a970d // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/zerolog v1.31.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// HV2 related packages
replace (
	cosmossdk.io/errors => github.com/0xPolygon/cosmos-sdk/errors v1.0.0
	cosmossdk.io/log => github.com/0xPolygon/cosmos-sdk/log v1.2.1
	cosmossdk.io/math => github.com/0xPolygon/cosmos-sdk/math v1.2.0

	github.com/cometbft/cometbft => github.com/0xPolygon/cometbft v0.1.0-beta

	github.com/ethereum/go-ethereum => github.com/maticnetwork/bor v1.4.0
)
