all:
	$(info =================== build vm manager =================)
	cd src && go build -o ../bin/vm_manager -v

list:
	PACKAGES=`go list ./... | grep -v /vendor/`
	VETPACKAGES=`go list ./... | grep -v /vendor/ | grep -v /examples/`
	GOFILES=`find . -name "*.go" -type f -not -path "./vendor/*"`

	@echo ${PACKAGES}
	@echo ${VETPACKAGES}
	@echo ${GOFILES}

fmt:
	$(info =================== format code ======================)
	@gofmt -s -w ${GOFILES}

fmt-check:
	$(info =================== format check =====================)
	@diff=$$(gofmt -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
			echo "Please run 'make fmt' and commit the result:"; \
			echo "$${diff}"; \
			exit 1; \
	fi;

test:
	@go test -cpu=1,2,4 -v -tags integration ./...

install:
	@govendor sync -v

docker:
	@docker build -t luotang/example:latest .

clean:
	$(info =================== build clean ======================)
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean

