GO111MODULE=on
YARN ?= yarn
GO ?= go

.PHONY: build
build: gocron node

.PHONY: build-race
build-race: enable-race build

.PHONY: run
run: build kill
	./bin/gocron-node &
	./bin/gocron web -e dev

.PHONY: run-race
run-race: enable-race run

.PHONY: kill
kill:
	-killall gocron-node

.PHONY: gocron
gocron:
	mkdir -p bin
	$(GO) build $(RACE) -o bin/gocron ./cmd/gocron

.PHONY: node
node:
	mkdir -p bin
	$(GO) build $(RACE) -o bin/gocron-node ./cmd/node

.PHONY: test
test:
	$(GO) test $(RACE) ./...

.PHONY: test-race
test-race: enable-race test

.PHONY: enable-race
enable-race:
	$(eval RACE = -race)

.PHONY: package
package: build-vue statik
	bash ./package.sh

.PHONY: package-all
package-all: build-vue statik
	bash ./package.sh -p 'linux darwin windows'

.PHONY: build-vue
build-vue:
	cd web/vue && $(YARN) run build
	rm -rf web/public
	mkdir -p web/public
	cp -r web/vue/dist/* web/public/

.PHONY: install-vue
install-vue:
	cd web/vue && $(YARN) install --frozen-lockfile

.PHONY: run-vue
run-vue:
	cd web/vue && $(YARN) run dev

.PHONY: statik
statik:
	$(GO) run github.com/rakyll/statik -src=web/public -dest=internal -f

.PHONY: lint
lint:
	golangci-lint run
	cd web/vue && $(YARN) run lint

.PHONY: clean
clean:
	rm -f bin/gocron bin/gocron.exe
	rm -f bin/gocron-node bin/gocron-node.exe
	rm -rf gocron-package gocron-node-package
