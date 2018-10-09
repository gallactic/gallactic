GOTOOLS = \
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v2

PACKAGES=$(shell go list ./... | grep -v '/vendor/')
INCLUDE = -I=. -I=${GOPATH}/src
BUILD_TAGS?=gallactic
BUILD_FLAGS = -ldflags "-X github.com/gallactic/gallactic/version.GitCommit=`git rev-parse --short=8 HEAD`"
SPUTNIKVM_PATH = $(GOPATH)/src/github.com/gallactic/sputnikvm-ffi

all: tools deps build test install


########################################
### Build Gallactic
build:
	CGO_LDFLAGS="$(SPUTNIKVM_PATH)/c/libsputnikvm.a -ldl" go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/gallactic ./cmd/gallactic/

build_race:
	CGO_LDFLAGS="$(SPUTNIKVM_PATH)/c/libsputnikvm.a -ldl" go build -race $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/gallactic ./cmd/gallactic

install:
	CGO_LDFLAGS="$(SPUTNIKVM_PATH)/c/libsputnikvm.a -ldl" go install $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' ./cmd/gallactic


########################################
### Docker
docker:
	docker build . --tag gallactic

########################################
### Tools & dependencies

tools:
	@cargo --version || (echo "Install Rust first; see https://rustup.rs/"; false)
	@echo "Installing tools"
	go get -u -v $(GOTOOLS)
	@gometalinter.v2 --install


deps:
	@rm -rf vendor/
	@echo "Running dep"
	@dep ensure -v
	@echo "Building Sputnikvm Library..."
	rm -rf $(SPUTNIKVM_PATH)
	mkdir $(SPUTNIKVM_PATH)
	cd $(SPUTNIKVM_PATH) && git clone https://github.com/gallactic/sputnikvm-ffi.git .
	cd $(SPUTNIKVM_PATH)/c && make build

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


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build build_race install docker
.PHONY: test test_race test_release test100
.PHONY: tools deps
.PHONY: fmt metalinter