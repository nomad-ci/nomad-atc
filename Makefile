build:
	GOOS=linux go build -o bin/driver driver/main.go
	cd bin; tar czvf driver.tar.gz driver
	cd bin; go-bindata -o ../assets/bindata.go -pkg assets driver.tar.gz
	go build -o bin/atc ./cmd/atc

test-atc:
	go build -o bin/atc ./cmd/atc
	bin/atc --no-really-i-dont-want-any-auth --log-level=info --external-url=http://10.0.1.168:8080 --nomad-datacenters=phx --nomad-url=http://172.23.160.3:4646

test-run:
	GOOS=linux go build -o bin/driver driver/main.go
	cd bin; tar czvf driver.tar.gz driver
	cd bin; go-bindata -o ../assets/bindata.go -pkg assets driver.tar.gz
	go run cmd/atc/main.go --no-really-i-dont-want-any-auth --log-level=info --external-url=http://10.0.1.168:8080 --nomad-datacenters=phx --nomad-url=http://172.23.160.3:4646

test-debug:
	GOOS=linux go build -o bin/driver driver/main.go
	cd bin; tar czvf driver.tar.gz driver
	cd bin; go-bindata -o ../assets/bindata.go -pkg assets driver.tar.gz
	dlv debug ./cmd/atc -- --no-really-i-dont-want-any-auth --log-level=debug
