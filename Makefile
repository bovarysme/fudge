APP=fudge
OUTPUT=build

.PHONY: all clean generate linux-amd64 openbsd-amd64

all: linux-amd64 openbsd-amd64

generate:
	go generate

linux-amd64: generate
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT)/$(APP)-linux-amd64 main.go

openbsd-amd64: generate
	GOOS=openbsd GOARCH=amd64 go build -o $(OUTPUT)/$(APP)-openbsd-amd64 main.go

clean:
	rm -rf $(OUTPUT)/$(APP)-*
