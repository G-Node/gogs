LDFLAGS += -X "github.com/G-Node/gogs/pkg/setting.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/G-Node/gogs/pkg/setting.BuildGitHash=$(shell git rev-parse HEAD)"

DATA_FILES := $(shell find conf | sed 's/ /\\ /g')
LESS_FILES := $(wildcard public/less/gogs.less public/less/_*.less)
GENERATED  := pkg/bindata/bindata.go public/css/gogs.css

OS := $(shell uname)

TAGS = ""
BUILD_FLAGS = "-v"

RELEASE_ROOT = "release"
RELEASE_GOGS = "release/gogs"
NOW = $(shell date -u '+%Y%m%d%I%M%S')
GOVET = go tool vet -composites=false -methods=false -structtags=false

.PHONY: build pack release bindata clean

.IGNORE: public/css/gogs.css

all: build

check: test

dist: release

web: build
	./gogs web

govet:
	$(GOVET) gogs.go
	$(GOVET) models pkg routes

build: $(GENERATED)
	go install $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)'
	cp '$(GOPATH)/bin/gogs' .

build-dev: $(GENERATED) govet
	go install $(BUILD_FLAGS) -tags '$(TAGS)'
	cp '$(GOPATH)/bin/gogs' .

build-dev-race: $(GENERATED) govet
	go install $(BUILD_FLAGS) -race -tags '$(TAGS)'
	cp '$(GOPATH)/bin/gogs' .

pack:
	rm -rf $(RELEASE_GOGS)
	mkdir -p $(RELEASE_GOGS)
	cp -r gogs LICENSE README.md README_ZH.md templates public scripts $(RELEASE_GOGS)
	rm -rf $(RELEASE_GOGS)/public/config.codekit $(RELEASE_GOGS)/public/less
	cd $(RELEASE_ROOT) && zip -r gogs.$(NOW).zip "gogs"

release: build pack

bindata: pkg/bindata/bindata.go

pkg/bindata/bindata.go: $(DATA_FILES)
	go-bindata -o=$@ -ignore="\\.DS_Store|README.md|TRANSLATORS|auth.d" -pkg=bindata conf/...

less: public/css/gogs.css

public/css/gogs.css: $(LESS_FILES)
	@type lessc >/dev/null 2>&1 && lessc $< >$@ || echo "lessc command not found, skipped."

clean:
	go clean -i ./...

clean-mac: clean
	find . -name ".DS_Store" -print0 | xargs -0 rm

test:
	go test -cover -race ./...

fixme:
	grep -rnw "FIXME" cmd routers models pkg

todo:
	grep -rnw "TODO" cmd routers models pkg

# Legacy code should be remove by the time of release
legacy:
	grep -rnw "LEGACY" cmd routes models pkg
