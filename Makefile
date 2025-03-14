BIN=go
OUTPATH=./bin


.PHONY: build test watch-test bench watch-bench coverage tools lint lint-fix audit outdated weight latest proto-all copy_html
build: create_build_folder copy_env copy_html
	${BIN} build -v -o ${OUTPATH} ./...

create_build_folder:
	@if [ -d "${OUTPATH}" ]; then find "${OUTPATH}" -mindepth 1 -delete; fi
	@if [ ! -d "${OUTPATH}" ]; then mkdir -p "${OUTPATH}"; fi

copy_env:
	@envfile=$$(find . -name ".env" -print -quit); \
	echo "Found .env file at $${envfile}"; \
	if [ -f "$${envfile}" ] && [ -f "${OUTPATH}/.env" ]; then \
		echo "Removing existing .env file from ${OUTPATH}"; \
		rm -f "${OUTPATH}/.env" || exit 1; \
	fi; \
	if [ -f "$${envfile}" ]; then \
		echo "Copying .env file from $${envfile} to ${OUTPATH}"; \
		cp -f "$${envfile}" "${OUTPATH}/" || exit 1; \
	else \
		echo "No .env file found to copy"; \
	fi

copy_html:
	@htmlfiles=$$(find . -name "*.html"); \
	if [ -z "$${htmlfiles}" ]; then \
		echo "No .html files found to copy"; \
	else \
		for htmlfile in $${htmlfiles}; do \
			echo "Copying $${htmlfile} to ${OUTPATH}"; \
			cp -f "$${htmlfile}" "${OUTPATH}/" || exit 1; \
		done; \
	fi

test:
	go test -race -v ./...
watch-test:
	reflex -t 50ms -s -- sh -c 'go test -race -v ./...'

bench:
	go test -benchmem -count 3 -bench ./...
watch-bench:
	reflex -t 50ms -s -- sh -c 'go test -benchmem -count 3 -bench ./...'

coverage:
	${BIN} test -v ./... -coverprofile=cover.out -covermode=atomic
	${BIN} tool cover -html=cover.out -o cover.html

tools:
	${BIN} install github.com/cespare/reflex@latest
	${BIN} install github.com/rakyll/gotest@latest
	${BIN} install github.com/psampaz/go-mod-outdated@latest
	${BIN} install github.com/jondot/goweight@latest
	${BIN} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${BIN} get -t -u golang.org/x/tools/cmd/cover
	${BIN} install github.com/sonatype-nexus-community/nancy@latest
	go mod tidy

lint:
	golangci-lint run --timeout 60s --max-same-issues 50 ./...
lint-fix:
	golangci-lint run --timeout 60s --max-same-issues 50 --fix ./...

audit:
	${BIN} list -json -m all | nancy sleuth

outdated:
	${BIN} list -u -m -json all | go-mod-outdated -update -direct

weight:
	goweight

latest:
	${BIN} get -t -u ./...
