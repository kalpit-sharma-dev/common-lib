PKG_LIST := $(shell go list ./... | sed -E 's;gitlab.kksharmadevdev.com/platform/platform-common-lib/src/[a-z][0-9]*/(.*);./\1;' | grep -vE 'vendor/|testApps/jsonMerge')
PKG_LIST_GOMOD = $(shell find . \( -path ./vendor -o -path ./testApps -o -path ./jsonMerge \) -prune -false -o -type f -name 'go.mod' | sed -e 's/\/go.mod//g')
PKG_LIST_COVERAGE := $(shell echo "${PKG_LIST}" | tr ' ' '\n' | grep -vE '/mock|/mocks|mock/|mocks/|Mock/|Mocks/|testApps|/example|/examples')

.SILENT: list list-coverage

all: dependencies build-all test cover

list:
	echo ${PKG_LIST}

list-coverage:
	echo ${PKG_LIST_COVERAGE}

build-all: app-build
	go install -v ${PKG_LIST}

dependencies:
	rm -rf vendor
	go mod download

api-windows-pdh-build:
	GOOS=windows go install ./api/win/pdh

app-build:
	cp ./testApps/app/versioninfo.json ./app/generate/versioninfo.json
	go generate ./app/generate
	go build -o ./testApps/app/application ./testApps/app/.; \
	rm ./app/generate/versioninfo.json; \
	rm ./testApps/app/application; \


cassandra-build:
	go install ./cassandra
	go install ./cassandra/mock
	go install ./cassandra/cql
	go install ./cassandra/cql/mock

checksum-build:
	go install ./checksum
	go install ./checksum/mock

circuit-build:
	go install ./circuit
	go install ./circuit/example

clar-build:
	go install ./clar
	go install ./clar/mock

config-build:
	go install ./config
	go install ./config/mock

communication-build:
	go install ./communication/udp
	go install ./communication/udp/mock
	go install ./communication/udp/example
	go install ./communication/http
	go install ./communication/http/client

db-build:
	go install ./db
	go install ./db/mock
	go install ./db/example
	go install ./db/mssql
	go install ./db/postgresql
	go install ./db/sqlite/mock

db-sqlite-build:
#	go install ./db/sqlite

distributed-build:
	go install ./distributed
	go install ./distributed/leader-election
	go install ./distributed/lock
	go install ./distributed/queue
	go install ./distributed/scheduler
	go install ./distributed/zookeeper

downloader-build:
	go install ./downloader
	go install ./downloader/http
	go install ./downloader/http/example
	go install ./downloader/grab
	go install ./downloader/grab/example
	go install ./downloader/mock

encoding-build:
	go install ./encoding/aes

entitlement-build:
	go install ./entitlement
	go install ./entitlement/mock

env-build:
	go install ./env
	go install ./env/mock

exception-build:
	go install ./exception

exec-build:
	go install ./exec
	go install ./exec/mock
	go install ./exec/with

freecache-build:
	go install ./freecache

jobprocessor-build:
	go install ./jobprocessor

json-build:
	go install ./json
	go install ./json/mock

kafka-build:
	go install ./kafka
	go install ./kafka/encode
	go install ./kafka/mock

cryptomgr-build:
	go install ./cryptomgr
	go install ./cryptomgr/example
	go install ./cryptomgr/mock

messaging-build:
	go install ./messaging
	go install ./messaging
	go install ./messaging/mock
	go install ./messaging/consumer
	go install ./messaging/consumer/examples/consumer
	go install ./messaging/producer
	go install ./messaging/producer/examples/async_producer
	go install ./messaging/producer/examples/mass_producer
	go install ./messaging/producer/examples/sync_producer
	go install ./messaging/producer/examples/timed_producer

metric-build:
	go install ./metric
	go install ./metric/example

namedpipes-build:
	go install ./namedpipes
	go install ./namedpipes/mock
	GOOS=windows go install ./namedpipes/npipe
	GOOS=windows go install ./namedpipes/winio

plugin-build:
	go install ./plugin/protocol
	go install ./plugin/protocol/mock
	go install ./plugin/protocol/http
	GOOS=windows go install ./plugin/wmi
	GOOS=windows go install ./plugin/wmi/wmiMock

pluginUtils-build:
	go install ./pluginUtils
	go install ./pluginUtils/mock

procParser-build:
	go install ./procParser
	go install ./procParser/mock

ratelimit-build:
	go install ./ratelimit
	go install ./ratelimit/mock

redis-client-build:
	go install ./redis
	go install ./redis/redismock

filter-build:
	go install ./filter
	go install ./filter/command
	go install ./filter/converters
	go install ./filter/mock

retry-build:
	go install ./retry

runtime-build:
	go install ./runtime/logger
	go install ./runtime/logger/mocks
	go install ./runtime/logger/example
	go install ./runtime/util
	go install ./runtime/audit
	go install ./runtime/audit/mocks
	go install ./runtime/audit/example

sanitize-build:
	go install ./sanitize

servicemanager-build:
	go install ./servicemanager
	GOOS=darwin go install ./servicemanager
	GOOS=windows go install ./servicemanager

services-build:
	go install ./services
	go install ./services/mock
	go install ./services/model

setupUtils-build:
	go install ./setupUtils
	go install ./setupUtils/appManagers

snowflake-build:
	go install ./snowflake

sync-build:
	go install ./sync
	go install ./sync/mock
	go install ./sync/zookeeper
	go install ./sync/zookeeper/mock

testApps-build:
#	for file in ./testApps/* ; do go install $$file; done

timeutil-build:
	go install ./timeutil

user-build:
	go install ./user

utils-build:
	go install ./utils

uuid-build:
	go install ./uuid

web-build:
	go install ./web
	go install ./web/mock
	go install ./web/microService
	go install ./web/rest
	go install ./web/filter
	go install ./web/filter/strategies/tokenize
	go install ./web/pagination

webClient-build:
	go install ./webClient
	go install ./webClient/mock

integration-test:
	go test -tags=integration -coverprofile ./cassandra/cover.out -covermode=count ./cassandra

mock-all:
	go generate -run mockgen ${PKG_LIST}

check-mocks: app-build
	@git status -sb > /tmp/mocks-${CI_JOB_ID}-before
	make mock-all
	@git status -sb > /tmp/mocks-${CI_JOB_ID}-after
	@diff -U0 /tmp/mocks-${CI_JOB_ID}-before /tmp/mocks-${CI_JOB_ID}-after

test: app-build
	go test ${PKG_LIST}

test-gomod:
	RET=0 && for package in ${PKG_LIST_GOMOD}; do cd "${PWD}/$${package}"; env GO111MODULE=on go test . ; RET=$$(($${RET}+$${?})); done; exit $${RET}

junit-test: app-build
	go test ${PKG_LIST} -v | tee /dev/stderr | go-junit-report -set-exit-code > ./junit-report.xml

cover: app-build
	for package in ${PKG_LIST_COVERAGE}; do go test -covermode=count -coverprofile "./$${package}/cover.out" "./$${package}" ; done
	echo "mode: count" > cover-all.out
	for package in ${PKG_LIST_COVERAGE}; do tail -n +2 ./$${package}/cover.out >> cover-all.out ; done
	gocover-cobertura < cover-all.out > cover-cobertura.xml

# run cover before show
.PHONY: show
show:
	echo "Launching web browser to show overall coverage..."
	go tool cover -html=cover-all.out

.PHONY: code-quality
code-quality:
	golangci-lint run --timeout 300s -c quality.yaml ${PKG_LIST} > static-analysis.xml

code-quality-ci: app-build
	golangci-lint run --timeout 300s -c quality.yaml --issues-exit-code 0 --out-format code-climate ${PKG_LIST} > static-analysis.json

