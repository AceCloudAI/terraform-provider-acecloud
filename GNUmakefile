default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o terraform-provider-acecloud

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/AceCloudAI/acecloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	mv terraform-provider-acecloud ~/.terraform.d/plugins/registry.terraform.io/AceCloudAI/acecloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

.PHONY: test
test:
	CGO_ENABLED=0 go test -v -cover -timeout=120s -parallel=4 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 CGO_ENABLED=0 go test -v -cover -timeout=120m ./...

# --- E2E test targets with report generation ---

.PHONY: testacc-report
testacc-report: ## Run all acceptance tests with report
	@./scripts/run-tests.sh ./internal/...

.PHONY: testacc-vpc
testacc-vpc: ## Run VPC tests with report
	@./scripts/run-tests.sh ./internal/resources/vpc/

.PHONY: testacc-subnet
testacc-subnet: ## Run Subnet tests with report
	@./scripts/run-tests.sh ./internal/resources/subnet/

.PHONY: testacc-security-group
testacc-security-group: ## Run Security Group tests with report
	@./scripts/run-tests.sh ./internal/resources/security_group/

.PHONY: testacc-key-pair
testacc-key-pair: ## Run Key Pair tests with report
	@./scripts/run-tests.sh ./internal/resources/key_pair/

.PHONY: testacc-instance
testacc-instance: ## Run Instance tests with report
	@./scripts/run-tests.sh ./internal/resources/instance/

.PHONY: testacc-volume
testacc-volume: ## Run Volume tests with report
	@./scripts/run-tests.sh ./internal/resources/volume/

.PHONY: testacc-networking
testacc-networking: ## Run all networking tests (floating IP, router, router interface) with report
	@./scripts/run-tests.sh ./internal/resources/floating_ip/ ./internal/resources/floating_ip_association/ ./internal/resources/router/ ./internal/resources/router_interface/

.PHONY: testacc-lb
testacc-lb: ## Run all load balancer tests with report
	@./scripts/run-tests.sh ./internal/resources/load_balancer/ ./internal/resources/lb_listener/ ./internal/resources/lb_pool/ ./internal/resources/lb_pool_member/ ./internal/resources/lb_health_monitor/

.PHONY: testacc-autoscaling
testacc-autoscaling: ## Run auto-scaling tests with report
	@./scripts/run-tests.sh ./internal/resources/auto_scaling_template/ ./internal/resources/auto_scaling_deployment/

.PHONY: testacc-datasources
testacc-datasources: ## Run all data source tests with report
	@./scripts/run-tests.sh ./internal/datasources/...

.PHONY: testacc-phase1
testacc-phase1: ## Run Phase 1 (foundation) tests: VPC, Subnet, SG, Key Pair
	@REPORT_NAME=phase1-foundation ./scripts/run-tests.sh ./internal/resources/vpc/ ./internal/resources/subnet/ ./internal/resources/security_group/ ./internal/resources/key_pair/

.PHONY: testacc-phase2
testacc-phase2: ## Run Phase 2 (compute/storage) tests: Instance, Volume, Snapshot, Backup
	@REPORT_NAME=phase2-compute-storage ./scripts/run-tests.sh ./internal/resources/instance/ ./internal/resources/volume/ ./internal/resources/volume_attachment/ ./internal/resources/snapshot/ ./internal/resources/volume_backup/

.PHONY: testacc-phase3
testacc-phase3: ## Run Phase 3 (networking) tests: Floating IP, Router
	@REPORT_NAME=phase3-networking ./scripts/run-tests.sh ./internal/resources/floating_ip/ ./internal/resources/floating_ip_association/ ./internal/resources/router/ ./internal/resources/router_interface/

.PHONY: testacc-phase4
testacc-phase4: ## Run Phase 4 (load balancing) tests
	@REPORT_NAME=phase4-load-balancing ./scripts/run-tests.sh ./internal/resources/load_balancer/ ./internal/resources/lb_listener/ ./internal/resources/lb_pool/ ./internal/resources/lb_pool_member/ ./internal/resources/lb_health_monitor/

.PHONY: testacc-phase5
testacc-phase5: ## Run Phase 5 (auto-scaling) tests
	@REPORT_NAME=phase5-autoscaling ./scripts/run-tests.sh ./internal/resources/auto_scaling_template/ ./internal/resources/auto_scaling_deployment/

.PHONY: testacc-phase6
testacc-phase6: ## Run Phase 6 (IAM) tests: API Key
	@REPORT_NAME=phase6-iam ./scripts/run-tests.sh ./internal/resources/api_key/

.PHONY: testacc-phase7
testacc-phase7: ## Run Phase 7 (data sources) tests
	@REPORT_NAME=phase7-datasources ./scripts/run-tests.sh ./internal/datasources/...

.PHONY: fmt
fmt:
	gofmt -s -w -e .

.PHONY: lint
lint:
	golangci-lint run ./... --timeout=5m

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
		--provider-dir . \
		--provider-name acecloud

.PHONY: docs-validate
docs-validate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate \
		--provider-dir . \
		--provider-name acecloud

.PHONY: clean
clean:
	rm -f terraform-provider-acecloud terraform-provider-acecloud_v*
	rm -rf dist/

.PHONY: testacc-clean
testacc-clean: ## Remove all test reports
	rm -rf test-reports/

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Build & Install:"
	@echo "    build              Build the provider binary"
	@echo "    install            Build and install locally for terraform dev_overrides"
	@echo "    clean              Remove built binaries and dist/"
	@echo ""
	@echo "  Code Quality:"
	@echo "    fmt                Format source code"
	@echo "    lint               Run golangci-lint"
	@echo "    vet                Run go vet"
	@echo "    tidy               Tidy go.mod"
	@echo ""
	@echo "  Unit Tests:"
	@echo "    test               Run unit tests"
	@echo ""
	@echo "  Acceptance Tests (require: source .env.test):"
	@echo "    testacc            Run all acceptance tests (no report)"
	@echo "    testacc-report     Run all acceptance tests with report"
	@echo ""
	@echo "  Single Resource Tests (with report):"
	@echo "    testacc-vpc             VPC"
	@echo "    testacc-subnet          Subnet"
	@echo "    testacc-security-group  Security Group"
	@echo "    testacc-key-pair        Key Pair"
	@echo "    testacc-instance        Instance"
	@echo "    testacc-volume          Volume"
	@echo "    testacc-networking      Floating IP, Router, Router Interface"
	@echo "    testacc-lb              Load Balancer stack"
	@echo "    testacc-autoscaling     Auto-Scaling Template + Deployment"
	@echo "    testacc-datasources     All data sources"
	@echo ""
	@echo "  Phase Tests (with report):"
	@echo "    testacc-phase1     Foundation (VPC, Subnet, SG, Key Pair)"
	@echo "    testacc-phase2     Compute & Storage (Instance, Volume, Snapshot)"
	@echo "    testacc-phase3     Networking (Floating IP, Router)"
	@echo "    testacc-phase4     Load Balancing (LB, Listener, Pool, Member, Monitor)"
	@echo "    testacc-phase5     Auto-Scaling (Template, Deployment)"
	@echo "    testacc-phase6     IAM (API Key)"
	@echo "    testacc-phase7     Data Sources"
	@echo ""
	@echo "  Reports:"
	@echo "    testacc-clean      Remove all test reports"
	@echo ""
	@echo "  Docs:"
	@echo "    docs               Generate provider documentation via tfplugindocs"
	@echo "    docs-validate      Validate generated docs"
