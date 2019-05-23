default:

frontend:
	docker run --rm -i \
		-v "$(CURDIR):$(CURDIR)" \
		-w "$(CURDIR)/src" \
		node:alpine \
		sh -c "npm ci && npm run build && chown -R $(shell id -u):$(shell id -g) ../frontend node_modules"

pack: frontend
	go-bindata -o assets.go frontend/...

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

.PHONY: frontend
