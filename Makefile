NAME := kthxbye

GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(shell go env GOPATH)/bin
endif

word-split = $(word $2,$(subst -, ,$1))
cc-%: go.mod go.sum cmd/kthxbye/*.go
	$(eval GOOS := $(call word-split,$*,1))
	$(eval GOARCH := $(call word-split,$*,2))
	$(eval GOARM := $(call word-split,$*,3))
	$(eval GOARMBIN := $(patsubst %,v%,$(GOARM)))
	$(eval GOEXT := $(patsubst $(GOOS),,$(patsubst windows,.exe,$(GOOS))))
	$(eval BINARY := "$(NAME)-$(GOOS)-$(GOARCH)$(GOARMBIN)$(GOEXT)")
	@awk -v bin=$(BINARY) -v goos=$(GOOS) -v goarch=$(GOARCH) -v goarm=$(GOARM) 'BEGIN { printf "[+] %-25s GOOS=%-10s GOARCH=%-10s GOARM=%1s\n", bin, goos, goarch, goarm }'
	@env CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -o $(BINARY) ./cmd/$(NAME)
	@test -f $(BINARY)
	@awk -v bin=$(BINARY) -v size=`du -h $(BINARY) | awk '{print $$1}'` -v type="`file -b $(BINARY)`" 'BEGIN { printf "[-] %-25s SIZE=%-10s TYPE=%s\n", bin, size, type }'

PLATFORMS := cc-darwin-amd64 cc-darwin-arm64 cc-dragonfly-amd64 cc-freebsd-386 cc-freebsd-amd64 cc-freebsd-arm-5 cc-freebsd-arm-6 cc-freebsd-arm-7 cc-linux-386 cc-linux-amd64 cc-linux-arm-5 cc-linux-arm-6 cc-linux-arm-7 cc-linux-arm64 cc-linux-ppc64 cc-linux-ppc64le cc-linux-mips cc-linux-mipsle cc-linux-mips64 cc-linux-mips64le cc-linux-s390x cc-netbsd-386 cc-netbsd-amd64 cc-netbsd-arm-5 cc-netbsd-arm-6 cc-netbsd-arm-7 cc-openbsd-386 cc-openbsd-amd64 cc-openbsd-arm-5 cc-openbsd-arm-6 cc-openbsd-arm-7 cc-solaris-amd64 cc-windows-386 cc-windows-amd64 cc-windows-arm64
crosscompile: $(PLATFORMS)
	@echo

.PHONY: clean
clean:
	rm -f $(NAME) $(NAME)-*

.DEFAULT_GOAL := $(NAME)
$(NAME): go.mod go.sum cmd/kthxbye/*.go
	go build ./cmd/$(NAME)

.PHONY: tools-go-mod-tidy
tools-go-mod-tidy:
	@for f in $(wildcard tools/*/go.mod) ; do echo ">>> $$f" && cd $(CURDIR)/`dirname "$$f"` && go mod tidy && cd $(CURDIR) ; done

$(GOBIN)/golangci-lint: tools/golangci-lint/go.mod tools/golangci-lint/go.sum
	go install -modfile=tools/golangci-lint/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
.PHONY: lint
lint: $(GOBIN)/golangci-lint
	$(ENV) golangci-lint run -v --timeout 5m -E staticcheck,misspell,promlinter,revive,tenv,errorlint,exportloopref,predeclared
