GOPATH := $(shell go env GOPATH)

all: install

checks:
	@echo "Checking deps:"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)
	@(env bash $(PWD)/buildscripts/checkgopath.sh)

getdeps: checks
	@echo "Installing golint:" && go get -u github.com/golang/lint/golint
	@echo "Installing gocyclo:" && go get -u github.com/fzipp/gocyclo
	@echo "Installing deadcode:" && go get -u github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell:" && go get -u github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign:" && go get -u github.com/gordonklaus/ineffassign
	@echo "Installing Simd:" && cd contrib/Simd && \
		cmake . -DCMAKE_INSTALL_PREFIX:PATH=/tmp/simd -DTOOLCHAIN="" -DTARGET="" && make -j8 install

verifiers: vet fmt lint cyclo spelling

vet:
	@echo "Running $@:"
	@go tool vet -all ./cmd
	@go tool vet -shadow=true ./cmd

fmt:
	@echo "Running $@:"
	@gofmt -s -l cmd

lint:
	@echo "Running $@:"
	@${GOPATH}/bin/golint -set_exit_status github.com/minio/xray...

ineffassign:
	@echo "Running $@:"
	@${GOPATH}/bin/ineffassign .

cyclo:
	@echo "Running $@:"
	@${GOPATH}/bin/gocyclo -over 30 cmd

build: getdeps verifiers

deadcode:
	@${GOPATH}/bin/deadcode

spelling:
	@${GOPATH}/bin/misspell -error `find cmd/`

gomake-all: build
	@echo "Installing xray:"
	@go install -v

pkg-add:
	${GOPATH}/bin/govendor add $(PKG)

pkg-update:
	${GOPATH}/bin/govendor update $(PKG)

pkg-remove:
	${GOPATH}/bin/govendor remove $(PKG)

pkg-list:
	@$(GOPATH)/bin/govendor list

install: gomake-all

clean:
	@echo "Cleaning up all the generated files:"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf release xray
