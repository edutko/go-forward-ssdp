all: forward-ssdp-freebsd-amd64 forward-ssdp-windows-amd64

forward-ssdp: out/forward-ssdp
forward-ssdp-freebsd-amd64: out/freebsd-amd64/forward-ssdp
forward-ssdp-windows-amd64: out/windows-amd64/forward-ssdp.exe

clean:
	rm -rf out

test:
	go test -race ./...

GO_INTERNAL_FILES=$(shell find internal -name '*.go')

out/forward-ssdp: cmd/forward-ssdp $(GO_INTERNAL_FILES) go.sum
	[ -d out/freebsd-amd64 ] || mkdir -p out/freebsd-amd64
	GOOS=freebsd GOARCH=amd64 go build -o $@ ./cmd/forward-ssdp

out/freebsd-amd64/forward-ssdp: cmd/forward-ssdp $(GO_INTERNAL_FILES) go.sum
	[ -d out/freebsd-amd64 ] || mkdir -p out/freebsd-amd64
	GOOS=freebsd GOARCH=amd64 go build -o $@ ./cmd/forward-ssdp

out/windows-amd64/forward-ssdp.exe: cmd/forward-ssdp $(GO_INTERNAL_FILES) go.sum
	[ -d out/windows-amd64 ] || mkdir -p out/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/forward-ssdp
