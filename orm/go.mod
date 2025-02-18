module cosmossdk.io/orm

go 1.23.2

require (
	cosmossdk.io/api v0.7.5
	cosmossdk.io/core v0.11.0
	cosmossdk.io/depinject v1.0.0
	cosmossdk.io/errors v1.0.0-beta.7
	github.com/cosmos/cosmos-db v1.0.2
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.6.0
	github.com/iancoleman/strcase v0.2.0
	github.com/regen-network/gocuke v0.6.2
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	google.golang.org/grpc v1.64.1
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/alecthomas/participle/v2 v2.0.0-alpha7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/apd/v3 v3.1.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/gogoproto v1.7.0 // indirect
	github.com/cucumber/common/gherkin/go/v22 v22.0.0 // indirect
	github.com/cucumber/common/messages/go/v17 v17.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/linxGnu/grocksdb v1.8.12 // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.47.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

// HV2 related packages
replace (
	cosmossdk.io/api => github.com/0xPolygon/cosmos-sdk/api v0.7.5
	cosmossdk.io/core => github.com/0xPolygon/cosmos-sdk/core v0.11.3-0.20241126102051-89dc71d02611
	cosmossdk.io/depinject => github.com/0xPolygon/cosmos-sdk/depinject v1.0.0
	cosmossdk.io/errors => github.com/0xPolygon/cosmos-sdk/errors v1.0.0-beta.7.0.20241126102051-89dc71d02611
)
