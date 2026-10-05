.PHONY: run test fmt vet eval clean

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

eval:
	go run ./cmd/evals

clean:
	rm -rf tmp/
