all: forward-ssdp-freebsd-amd64 forward-ssdp-windows-amd64

clean:
	rm -rf out

forward-ssdp-freebsd-amd64: out/freebsd-amd64/forward-ssdp cmd/forward-ssdp
out/freebsd-amd64/forward-ssdp cmd/forward-ssdp:
	[ -d out/freebsd-amd64 ] || mkdir -p out/freebsd-amd64
	GOOS=freebsd GOARCH=amd64 go build -o out/freebsd-amd64/forward-ssdp ./cmd/forward-ssdp

forward-ssdp-windows-amd64: out/windows-amd64/forward-ssdp.exe
out/windows-amd64/forward-ssdp.exe:
	[ -d out/windows-amd64 ] || mkdir -p out/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o out/windows-amd64/forward-ssdp.exe ./cmd/forward-ssdp
