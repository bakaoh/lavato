BIN = lavato
REVISION = $(shell git log | head -n 1 | cut  -f 2 -d ' ')

dev:
	go-bindata -dev -pkg assets -o assets/assets.go assets/public_html/...
	go run main.go regus

web:
	cd assets/public_html/ && browser-sync start -s -f . --no-notify --port 5678

clean:
	rm -f $(BIN)

build: clean
	go-bindata -pkg assets -o assets/assets.go assets/public_html/...
	go build -ldflags "-X main.revision=$(REVISION)"

rsync:
	rsync -avz -e 'ssh -p 443' lavato* root@128.199.206.130:/tserver/go-projects/lavato/

pulldb:
	rsync -chavzP -e 'ssh -p 443' root@128.199.206.130:/tserver/go-projects/lavato/regus.db .

pushdb:
	rsync -avz -e 'ssh -p 443' regus.db root@128.199.206.130:/tserver/go-projects/lavato/
