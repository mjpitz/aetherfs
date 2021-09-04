SKAFFOLD_DEFAULT_REPO ?= ghcr.io/mjpitz
CWD = $(shell pwd)
VERSION ?= latest

define HELP_TEXT
Welcome to aetherfs!

Targets:
  help             provides help text
  dist             recompiles aetherfs binaries
  docker           rebuild the aetherfs docker container
  docker/devtools  rebuild docker container containing developer tools
  gen              regenerate the API code from protocol buffers
  legal            prepends legal header to source code
  release          releases aetherfs

endef
export HELP_TEXT

help:
	@echo "$$HELP_TEXT"

legal: .legal
.legal:
	addlicense -f ./legal/header.txt -skip yaml .

docker/devtools: .docker/devtools
.docker/devtools:
	docker build ./docker/devtools -t $(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools

gen: .gen
.gen:
	@rm -rf api gen
	docker run --rm -it \
		-v $(CWD):/home \
		-w /home \
		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
		sh -c 'buf lint && buf generate'

dist: .dist
.dist:
	docker run --rm -it \
		-v $(CWD):/home \
		-w /home \
		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
		sh -c "goreleaser --snapshot --skip-publish --rm-dist"

docker: .docker
.docker:
	docker build . \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:$(VERSION) \
		--file ./docker/aetherfs/Dockerfile

shell:
	docker run --rm -it \
		-v $(CWD):/home \
		-w /home \
		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
		sh

# release - used to generate core release assets such as binaries and container images.

release:
	docker run --rm -it \
    		-v $(CWD):/home \
    		-w /home \
    		$(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools \
    		sh -c "goreleaser"

	docker buildx build . \
		--platform linux/amd64,linux/arm64 \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/aetherfs:$(VERSION) \
		--file ./docker/aetherfs/Dockerfile \
		--push

# proto-tar - invoked from goreleaser to produce a tar.gz of proto files for the version.
# todo: use bufs image concept

PROTO = $(PWD)/proto
DIST = $(PWD)/dist

.proto-tar:
	[[ -d "$(DIST)/aetherfs_proto" ]] || { \
		mkdir -p $(DIST)/aetherfs_proto; \
		cp -R $(PROTO)/* $(DIST)/aetherfs_proto; \
		tar -czf $(DIST)/aetherfs_proto.tar.gz -C $(DIST)/aetherfs_proto/ . ; \
	}
