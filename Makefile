shell:
	nix develop -c $$SHELL

server/build:
	./dev-scripts/build-server.sh

server/run: server/build
	./tmp/server

secrets/hmac:
	xxd -l32 /dev/urandom | xxd -r -ps | base64 | tr -d = | tr + - | tr / _
