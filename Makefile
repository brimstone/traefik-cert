ifdef DEV
traefik-cert: *.go cmd/*.go server/*.go
	go build -v -i
else
traefik-cert: *.go cmd/*.go server/*.go
	sleep 1
	docker build -t brimstone/traefik-cert --build-arg PACKAGE=github.com/brimstone/traefik-cert .
	cid="$$(docker create --name traefik-cert brimstone/traefik-cert)" \
	&& docker cp $$cid:/traefik-cert . \
	&& docker rm $$cid
endif

.PHONY: test
test: traefik-cert
	docker rm -vf cert || true
	sleep 3
	./traefik-cert getcert -u dev.sprinkle.cloud -d dev.sprinkle.cloud -j $(shell jwt gentoken -k private.key '{"cert": {"domains": ["dev.sprinkle.cloud"]}}')

.PHONY: watch
watch:
	find Makefile *.go */*.go | entr -c make test

.PHONY: clean
clean:
	rm -f traefik-cert

.PHONY: serverloop
serverloop:
	while true; do docker run --rm -it -v /tmp/acme:/acme --label=traefik.frontend.rule=Host:dev.sprinkle.cloud --name cert -v ${PWD}/public.key:/public.key brimstone/traefik-cert; sleep 1; done

.PHONY: traefik
traefik:
	docker run --rm --name traefik -it -v /tmp/acme:/acme -p 80:80 -p 443:443 -v /var/run/docker.sock:/var/run/docker.sock traefik --entryPoints='Name:https Address::443 TLS' --entryPoints='Name:http Address::80 Redirect.EntryPoint:https' --defaultEntryPoints='http,https' --acme.httpChallenge.entryPoint=http --acme.acmeLogging=true --acme.entryPoint=https --acme.storage=/acme/acme.json --acme.onhostrule --docker --docker.watch --loglevel=info --web

.PHONY: image
image:
	docker buildx build --pull --platform=linux/arm64,linux/amd64 --push -t brimstone/traefik-cert .
