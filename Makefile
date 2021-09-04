SKAFFOLD_DEFAULT_REPO ?= ghcr.io/mjpitz
CWD = $(shell pwd)

define HELP_TEXT
Welcome to aetherfs!

Targets:
  help             provides help text
  docker/devtools  rebuild docker container containing developer tools
  gen              regenerate the API code from protocol buffers

endef
export HELP_TEXT

help:
	@echo "$$HELP_TEXT"

docker/devtools: .docker/devtools
.docker/devtools:
	docker build ./docker/devtools -t $(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools

gen: .gen
.gen:
	@rm -rf api gen
	docker run --rm -it -v $(CWD):/home -w /home $(SKAFFOLD_DEFAULT_REPO)/aetherfs-devtools sh -c 'buf lint && buf generate'
