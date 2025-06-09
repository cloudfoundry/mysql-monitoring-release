module github.com/cloudfoundry/mysql-diag

go 1.23.0

toolchain go1.23.3

require (
	code.cloudfoundry.org/bytefmt v0.41.0
	code.cloudfoundry.org/tlsconfig v0.28.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/fatih/color v1.18.0
	github.com/go-sql-driver/mysql v1.9.2
	github.com/olekukonko/tablewriter v1.0.7
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250607225305-033d6d78b36a // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/square/certstrap v1.3.0 // indirect
	go.step.sm/crypto v0.66.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/olekukonko/tablewriter => github.com/olekukonko/tablewriter v0.0.5
