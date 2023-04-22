.DEFAULT_GOAL := help

# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
DEFAULT_BRANCH := main
THIS_FILE := $(lastword $(MAKEFILE_LIST))
PKG := github.com/natemarks/ccloud-admin
VERSION := 0.0.0
COMMIT := $(shell git rev-parse HEAD)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
CDIR = $(shell pwd)
EXECUTABLES := ccloud-delete
GOOS := linux
GOARCH := amd64

CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DEFAULT_BRANCH := main

help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

clean-venv: ## re-create virtual env
	rm -rf .venv
	python3 -m venv .venv
	( \
       source .venv/bin/activate; \
       pip install --upgrade pip setuptools; \
    )

${EXECUTABLES}:
	@for o in $(GOOS); do \
	  for a in $(GOARCH); do \
        env GOOS=$${o} GOARCH=$${a} \
        go build  -v -o build/$${o}/$${a}/$@ ${PKG}/cmd/$@; \
	  done \
    done ; \

build: git-status ${EXECUTABLES}
	rm -f build/current
	ln -s $(CDIR)/build/$(COMMIT) $(CDIR)/build/current

release: git-status build
	mkdir -p release/$(VERSION)
	@for o in $(GOOS); do \
	  for a in $(GOARCH); do \
        tar -C ./build/$(COMMIT)/$${o}/$${a} -czvf release/$(VERSION)/ccloud-admin_$(VERSION)_$${o}_$${a}.tar.gz . ; \
	  done \
    done ; \

test:
	@go test -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}


lint: ##  run golint
	( \
			 go install golang.org/x/lint/golint@latest; \
			 golint ./...; \
			 test -z "$$(golint ./...)"; \
    )

shellcheck: ## use black to format python files
	( \
			 git ls-files '*.sh' |  xargs shellcheck --format=gcc; \
    )

static: lint ## run fmt, vet, goimports, gocyclo
	( \
			 gofmt -w  -s .; \
			 test -z "$$(go vet ./...)"; \
			 go install golang.org/x/tools/cmd/goimports@latest; \
			 goimports -w .; \
			 go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; \
			 test -z "$$(gocyclo -over 25 .)"; \
			 go install honnef.co/go/tools/cmd/staticcheck@latest ; \
			 staticcheck ./... ; \
    )

clean:
	-@rm ${OUT} ${OUT}-v*


bump: clean-venv  ## bump version in main branch
ifeq ($(CURRENT_BRANCH), $(DEFAULT_BRANCH))
	( \
	   source .venv/bin/activate; \
	   pip install bump2version; \
	   bump2version $(part); \
	)
else
	@echo "UNABLE TO BUMP - not on Main branch"
	$(info Current Branch: $(CURRENT_BRANCH), main: $(DEFAULT_BRANCH))
endif


git-status: ## require status is clean so we can use undo_edits to put things back
	@status=$$(git status --porcelain); \
	if [ ! -z "$${status}" ]; \
	then \
		echo "Error - working directory is dirty. Commit those changes!"; \
		exit 1; \
	fi

rebase: git-status ## rebase current feature branch on to the default branch
	git fetch && git rebase origin/$(DEFAULT_BRANCH)

.PHONY: build release static upload vet lint fmt gocyclo goimports test