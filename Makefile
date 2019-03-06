.PHONY: all
all: get test

.PHONY: get
get:
	go get -u golang.org/x/lint/golint \
		golang.org/x/tools/cmd/stringer \
		golang.org/x/tools/cmd/goimports \
		github.com/securego/gosec/cmd/gosec/... \
		honnef.co/go/tools/...

.PHONY: lint
lint:
	golint .
	gofmt -s -l .
	goimports -l -local=github.com/travelaudience/ .
	staticcheck ./...
	gosec ./...

.PHONY: test
test:
	go test -race -coverprofile=cover.out -covermode=atomic -count=2 .

.PHONY: fix
fix:
	go mod tidy
	gofmt -s -w .
	goimports -w -local=github.com/travelaudience/ .

.PHONY: gen
gen:
	go generate
	java -jar ./tmp/plantuml.jar -verbose doc/class.puml -tsvg
	java -jar ./tmp/plantuml.jar -verbose doc/graph.puml -tsvg