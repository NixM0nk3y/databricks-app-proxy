# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

VERSION=1.0

# Output helpers
# --------------

TASK_DONE = echo "✓  $@ done"
TASK_BUILD = echo "🛠️  $@ done"

export GITHUB_RUN_ID ?= 0
export GITHUB_SHA ?= $(shell git rev-list -1 HEAD --abbrev-commit)
export BUILD_DATE=$(shell date -u '+%Y%m%d')

all: test tokenvendor/build

deps:
	go get -v  ./...
	
tokenvendor/build:
	$(GOBUILD) -ldflags " \
		-X tokenvendor/pkg/version.Version=${VERSION}.${GITHUB_RUN_ID} \
		-X tokenvendor/pkg/version.BuildHash=${GITHUB_SHA} \
		-X tokenvendor/pkg/version.BuildDate=${BUILD_DATE}" \
		-o ./tokenvendor -v ./cmd
	@$(TASK_BUILD)

test: 
	@$(GOTEST) -v ./...
	@$(TASK_DONE)

clean:
	@$(GOCLEAN)
	@rm -f ./tokenvendor
	@$(TASK_DONE)