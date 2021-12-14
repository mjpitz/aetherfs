SKAFFOLD_DEFAULT_REPO ?= ghcr.io/mjpitz
CWD = $(shell pwd)
VERSION ?= latest

define HELP_TEXT
Welcome to aetherfs!

Targets:
  help             provides help text

  docker           rebuild the aetherfs docker container
  docker/devtools  rebuild docker container containing developer tools
  docker/release   releases aetherfs (will likely move)
  docker/shell     spins up an interactive shell with all dev tools
  in-docker        run targets in docker (useful to avoid local deps)

  lint             lints the code base
  legal            prepends legal header to source code
  test             run tests
  gen              regenerate the API code from protocol buffers
  dist             recompiles aetherfs binaries

endef
export HELP_TEXT

help:
	@echo "$$HELP_TEXT"

docker: .docker
.docker:
	docker build . \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:$(VERSION) \
		--file ./deploy/docker/aetherfs/Dockerfile

docker/devtools: .docker/devtools
.docker/devtools:
	docker build ./deploy/docker/devtools -t $(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools

in-docker:
	docker run --rm -i \
		-e VERSION \
		-e GITHUB_TOKEN \
		-v $(CWD):/home \
		-w /home \
		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
		make $(TARGETS)

docker/shell: .docker/shell
.docker/shell:
	docker run --rm -it \
		-v $(CWD):/home \
		-w /home \
		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
		sh

docker/release:
	docker buildx build . \
		--platform linux/amd64,linux/arm64 \
		--label "org.opencontainers.image.source=https://github.com/mjpitz/aetherfs" \
		--label "org.opencontainers.image.version=$(VERSION)" \
		--label "org.opencontainers.image.licenses=AGPL-3.0-only" \
		--label "org.opencontainers.image.title=AetherFS" \
		--label "org.opencontainers.image.description=" \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:$(VERSION) \
		--file ./deploy/docker/aetherfs/Dockerfile \
		--push

# actual targets

lint:
	./scripts/lint.sh

docs: .docs
.docs:
	helm-docs -c deploy/charts/aetherfs-datasets --dry-run | prettier --parser markdown > deploy/charts/aetherfs-datasets/README.md
	helm-docs -c deploy/charts/aetherfs-hub --dry-run | prettier --parser markdown > deploy/charts/aetherfs-hub/README.md

legal: .legal
.legal:
	addlicense -f ./legal/header.txt -skip yaml -skip yml deploy/docker internal proto scripts web/public web/src

test:
	go test -v -race -cover -coverprofile=.coverprofile ./...

gen: .gen
.gen:
	./scripts/buf.sh

dist: .dist
.dist:
	./scripts/dist-web.sh
	make legal
	./scripts/dist-go.sh

# useful shortcuts for release

tag/release:
	npm version "$(shell date +%y.%m.0)"
	git push --follow-tags

tag/patch:
	npm version patch
	git push --follow-tags
