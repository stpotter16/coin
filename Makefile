shell:
	nix develop -c $$SHELL

server/build:
	./dev-scripts/build-server.sh

server/run: server/build
	./tmp/server

server/live:
	./dev-scripts/serve.sh

secrets/hmac:
	xxd -l32 /dev/urandom | xxd -r -ps | base64 | tr -d = | tr + - | tr / _
