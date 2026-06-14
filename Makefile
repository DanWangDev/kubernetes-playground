.PHONY: cluster-create cluster-delete cluster-reset cluster-status
.PHONY: build lint vet test test-cover
.PHONY: all

# ── Cluster lifecycle ─────────────────────────────────────────────

cluster-create:
	go run ./cmd/cluster-create/

cluster-delete:
	go run ./cmd/cluster-delete/

cluster-reset: cluster-delete cluster-create

cluster-status:
	go run ./cmd/cluster-status/

# ── Build & check ──────────────────────────────────────────────────

build:
	go build ./...

lint: vet

vet:
	go vet ./...

# ── Tests ──────────────────────────────────────────────────────────

test:
	go test ./test/... -v -count=1

test-cover:
	go test ./test/... -v -cover -coverprofile=coverage.out

# ── All exercises ──────────────────────────────────────────────────
# Usage: make exercise-01, make exercise-02, etc.

exercise-%:
	go run ./exercises/$*/ --step=false

# ── CI ─────────────────────────────────────────────────────────────

all: build vet test
