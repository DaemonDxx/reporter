set -e
export GOFLAGS="-mod=vendor"
export GOOS="linux"
export GOARCH="amd64"

go build -o bin/ cmd/reporter.go