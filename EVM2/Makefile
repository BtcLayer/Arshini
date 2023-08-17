all: etleneum runcall

etleneum: $(shell find . -name "*.go") static/bundle.js
	go build

runcall: runlua/runlua.go runlua/cmd/runcall/main.go
	cd runlua/cmd/runcall && CC=$$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ../../../runcall

static/bundle.js: $(shell find client)
	GITHUB_REPO=etleneum/database-dev ./node_modules/.bin/rollup -c

deploy_test: etleneum
	GITHUB_REPO=etleneum/database-dev ./node_modules/.bin/rollup -c
	CC=$$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./etleneum
	ssh root@turgot 'systemctl stop etleneum-test'
	scp etleneum turgot:etleneum-test/etleneum
	ssh root@turgot 'systemctl start etleneum-test'

deploy: etleneum
	PRODUCTION=true GITHUB_REPO=etleneum/database ./node_modules/.bin/rollup -c
	CC=$$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./etleneum
	rsync etleneum turgot:.lightning1/plugins/etleneum-new
	ssh turgot 'ln1 plugin stop etleneum; mv .lightning1/plugins/etleneum-new .lightning1/plugins/etleneum; ln1 plugin start $$HOME/.lightning1/plugins/etleneum'
