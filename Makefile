# VERSION := $(shell git describe --tags)
# BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

WATCH_ADDR := localhost:7999

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Redirect error output to a file, so we can show it in development mode.
STDERR := /tmp/.$(PROJECTNAME)-stderr.txt

# PID file will keep the process id of the server
PID := /tmp/.$(PROJECTNAME).pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent





## install: Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
install: go-get

ci: compile go-test-cover



## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/Error:/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2


## clean: Clean build files. Runs `go clean` internally.
clean:
	@-rm $(GOBIN)/$(PROJECTNAME) 2> /dev/null
	@-$(MAKE) go-clean


tdd:
	@bash -c "trap 'make tdd-stop' EXIT; $(MAKE) tdd-setup"
	@bash -c "$(MAKE) watch run='make clean compile tdd-exec'"

tdd-setup: 
	@-$(MAKE) clean compile
	@echo "  ℹ  build stat for $(PROJECTNAME) is available at http://$(WATCH_ADDR)"
	@-$(MAKE) tdd-exec


tdd-exec:
	echo "  🚀  Run tests ..."
	@$(MAKE) go-test 2>&1 
	echo "  ☑️  tests done"

	# @cat $(PID) | sed "/^/s/^/  \>  PID: /"

tdd-stop:
	@-touch $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

## watch: Run given command when code changes. e.g; make watch run="echo 'hey'"
watch:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) yolo -i '*/*.go' -i '*.go' -e vendor -e bin -a "$(WATCH_ADDR)" -c "$(run)"


go-compile: go-get go-build

go-test:
	@go test ./... -coverpkg=./... -timeout 10s
 
go-test-cover:
	@go test ./... -coverpkg=./... -coverprofile=coverage.txt  -timeout 30s
	@go tool cover -func=coverage.txt

go-build:
	@echo "  ⚙️  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  🛠  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  🔎  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  🗑  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

go-generate-mocks:
	@mockgen -source=./core/operator.go  -destination=./mock_core/operator_mocks_test.go -package=core_test Operator
	# @mockgen -source=./core/conn.go -destination=./mock_core/conn_mocks_test.go Conn 
