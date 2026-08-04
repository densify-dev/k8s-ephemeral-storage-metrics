LOCALBIN ?= $(CURDIR)/bin

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: fmt vet test-unit govulncheck gosec

fmt: ## Format Go code.
	go fmt ./...

vet: ## Run Go static analysis.
	go vet ./...

test-unit: ## Run unit tests; e2e tests require external Helm chart.
	go test ./pkg/... ./cmd/...

govulncheck: | $(LOCALBIN)
	test -s $(LOCALBIN)/govulncheck || GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(LOCALBIN)/govulncheck ./...

gosec: | $(LOCALBIN)
	test -s $(LOCALBIN)/gosec || GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@latest
	$(LOCALBIN)/gosec ./...
