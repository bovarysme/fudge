APP=fudge
OUTPUT=build

.PHONY: all checksum clean generate linux-amd64 openbsd-amd64

all: linux-amd64 openbsd-amd64 checksum

generate:
	go generate

linux-amd64: generate
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT)/$(APP)-$@ main.go
	tar --transform="flags=r;s|$(OUTPUT)/$(APP)-$@|fudge|" \
		-czf $(OUTPUT)/fudge-$@.tar.gz \
		$(OUTPUT)/$(APP)-$@ static/ template/

openbsd-amd64: generate
	GOOS=openbsd GOARCH=amd64 go build -o $(OUTPUT)/$(APP)-$@ main.go
	tar --transform="flags=r;s|$(OUTPUT)/$(APP)-$@|fudge|" \
		-czf $(OUTPUT)/fudge-$@.tar.gz \
		$(OUTPUT)/$(APP)-$@ static/ template/

checksum:
	cd $(OUTPUT) && sha256sum -b $(APP)-*.tar.gz > sha256sum.txt

clean:
	rm -rf $(OUTPUT)/$(APP)-* $(OUTPUT)/sha256sum.txt
