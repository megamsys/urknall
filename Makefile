GOPATH  := $(GOPATH):$(shell pwd)/../../../../

define HG_ERROR

FATAL: you need mercurial (hg) to download megamd dependencies.
       Check README.md for details


endef

define GIT_ERROR

FATAL: you need git to download megamd dependencies.
       Check README.md for details
endef

define BZR_ERROR

FATAL: you need bazaar (bzr) to download megamd dependencies.
       Check README.md for details
endef

.PHONY: all check-path get hg git bzr get-code test

all: check-path get test

build: check-path get _go_test _urknall

# It does not support GOPATH with multiple paths.
check-path:
ifndef GOPATH
	@echo "FATAL: you must declare GOPATH environment variable, for more"
	@echo "       details, please check README.md file and/or"
	@echo "       http://golang.org/cmd/go/#GOPATH_environment_variable"
	@exit 1
endif

	@exit 0

get: hg git bzr get-code godep

hg:
	$(if $(shell hg), , $(error $(HG_ERROR)))

git:
	$(if $(shell git), , $(error $(GIT_ERROR)))

bzr:
	$(if $(shell bzr), , $(error $(BZR_ERROR)))


get-code:
	go get $(GO_EXTRAFLAGS) -u -d -t -insecure ./...

godep:
	go get $(GO_EXTRAFLAGS) github.com/tools/godep
	godep restore ./...

build: check-path get _go_test _urknall

_go_test:
	go clean  ./...
	@go test $(PACKAGES)

test: _go_test 

vet:
	@go vet $(PACKAGES)

_install_deadcode: git
	go get $(GO_EXTRAFLAGS) github.com/remyoudompheng/go-misc/deadcode

deadcode: _install_deadcode
	@go list ./... | sed -e 's;github.com/megamsys/urknall/;;' | xargs deadcode

deadc0de: deadcode




