test-run:
	GOOS=linux go build -o bin/driver driver/main.go
	cd bin; tar czvf driver.tar.gz driver
	go run cmd/atc/main.go --no-really-i-dont-want-any-auth --log-level=info --external-url=http://10.0.1.168:8080 --nomad-datacenters=phx --nomad-url=http://172.23.160.3:4646

test-debug:
	GOOS=linux go build -o bin/driver driver/main.go
	cd bin; tar czvf driver.tar.gz driver
	dlv debug ./cmd/atc -- --no-really-i-dont-want-any-auth --log-level=debug
