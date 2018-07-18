GOTOOLS = \
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v2

PACKAGES=$(shell go list ./... | grep -v '/vendor/')
INCLUDE = -I=. -I=${GOPATH}/src
BUILD_TAGS?=gallactic
BUILD_FLAGS = -ldflags "-X github.com/gallactic/gallactic/version.GitCommit=`git rev-parse --short=8 HEAD`"

all: check build test install

check: check_tools ensure_deps


########################################
### Build Gallactic
build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/gallactic ./cmd/gallactic/

build_race:
	CGO_ENABLED=0 go build -race $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/gallactic ./cmd/gallactic

install:
	CGO_ENABLED=0 go install $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' ./cmd/gallactic

########################################
### Tools & dependencies

check_tools:
	@# https://stackoverflow.com/a/25668869
	@echo "Found tools: $(foreach tool,$(notdir $(GOTOOLS)),\
        $(if $(shell which $(tool)),$(tool),$(error "No $(tool) in PATH")))"

get_tools:
	@echo "--> Installing tools"
	go get -u -v $(GOTOOLS)
	@gometalinter.v2 --install

update_tools:
	@echo "--> Updating tools"
	@go get -u $(GOTOOLS)

#Run this from CI
get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running dep"
	@dep ensure -vendor-only


#Run this locally.
ensure_deps:
	@rm -rf vendor/
	@echo "--> Running dep"
	@dep ensure

########################################
### Testing


test_release:
	@go test -tags release $(PACKAGES)

test100:
	@for i in {1..100}; do make test; done


### go tests
test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_race:
	@echo "--> Running go test --race"
	@go test -v -race $(PACKAGES)


########################################
### Formatting, linting, and vetting

fmt:
	@go fmt ./...

metalinter:
	@echo "--> Running linter"
	@gometalinter.v2 --vendor --deadline=600s --disable-all  \
		--enable=deadcode \
		--enable=gosimple \
	 	--enable=misspell \
		--enable=safesql \
		./...
		#--enable=gas \
		#--enable=maligned \
		#--enable=dupl \
		#--enable=errcheck \
		#--enable=goconst \
		#--enable=gocyclo \
		#--enable=goimports \
		#--enable=golint \ <== comments on anything exported
		#--enable=gotype \
	 	#--enable=ineffassign \
	   	#--enable=interfacer \
	   	#--enable=megacheck \
	   	#--enable=staticcheck \
	   	#--enable=structcheck \
	   	#--enable=unconvert \
	   	#--enable=unparam \
		#--enable=unused \
	   	#--enable=varcheck \
		#--enable=vet \
		#--enable=vetshadow \

metalinter_all:
	@echo "--> Running linter (all)"
	gometalinter.v2 --vendor --deadline=600s --enable-all --disable=lll ./...


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: check build build_race install
.PHONY: test test_race test_release test100
.PHONY: check_tools get_tools update_tools get_vendor_deps ensure_deps
.PHONY: fmt metalinter metalinter_all