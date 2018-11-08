GOTOOLS = \
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v2

PACKAGES=$(shell go list ./... | grep -v '/vendor/')
SPUTNIKVM_PATH = $(GOPATH)/src/github.com/gallactic/sputnikvm-ffi
TAGS=-tags 'gallactic'
LDFLAGS= -ldflags "-X github.com/gallactic/gallactic/version.GitCommit=`git rev-parse --short=8 HEAD`"
CFLAGS=CGO_LDFLAGS="$(SPUTNIKVM_PATH)/c/libsputnikvm.a -ldl"


all: tools deps build install test test_release

########################################
### Tools & dependencies
tools:
	@cargo --version || (echo "Install Rust first; see https://rustup.rs/"; false)
	@echo "Installing tools"
	go get $(GOTOOLS)
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
### Build Gallactic
build:
	$(CFLAGS) go build $(LDFLAGS) $(TAGS) -o build/gallactic ./cmd/gallactic/

install:
	$(CFLAGS) go install $(LDFLAGS) $(TAGS) ./cmd/gallactic

########################################
### Testing
test:
	$(CFLAGS) go test $(PACKAGES)

test_release:
	$(CFLAGS) go test -tags release $(PACKAGES)

#race condirion
test_race:
	$(CFLAGS) go test -race $(PACKAGES)


########################################
### Docker
docker:
	docker build . --tag gallactic



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
.PHONY: build install docker test test_race test_release
.PHONY: tools deps
.PHONY: fmt metalinter