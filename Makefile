EXEC = bin/datalog.linux-amd64
TEST = bin/datalog.linux-amd64.test

.PHONY: build

build: $(EXEC) $(TEST)
	docker build -t rxaxm/datalog .

$(EXEC):
	GOOS=linux GOARCH=amd64 go build -o $(EXEC) ./cmd/datalog/*.go

$(TEST):
	GOOS=linux GOARCH=amd64 go test -c -o $(TEST)
