.PHONY: all
all: get test

.PHONY: get
get:
	go get -u github.com/golang/lint/golint \
		golang.org/x/tools/cmd/goimports \
		github.com/securego/gosec/cmd/gosec/... \
		honnef.co/go/tools

.PHONY: lint
lint:
	golint .
	gofmt -s -l .
	goimports -l -local=github.com/travelaudience/ .
	staticcheck ./...
	gosec ./...

.PHONY: test
test: lint
	go test -race -coverprofile=cover.out -covermode=atomic -count=2 .

.PHONY: fix
fix:
	go mod tidy
	gofmt -s -w .
	goimports -w -local=github.com/travelaudience/ .

.PHONY: gen
gen:
	java -jar ./tmp/plantuml.jar -verbose doc/class.puml -tsvg
	java -jar ./tmp/plantuml.jar -verbose doc/graph.puml -tsvg