ACTIVE_PLUGINS := osHealth
VERSION ?= devel

# Example usage:
# make all VERSION=1.0.0
# make build [PLUGIN=pluginName] VERSION=1.0.0
# make run [PLUGIN=pluginName] ARGS="-v" VERSION=1.0.0
# make send HOST=user@host [PLUGIN=pluginName]

.PHONY: clean all

all: clean linux-amd64-all linux-arm64-all windows-amd64-all freebsd-amd64-all

linux-amd64-all: GOOS=linux
linux-amd64-all: GOARCH=amd64
linux-amd64-all: TARGET=GOOS=$(GOOS) GOARCH=$(GOARCH)
linux-amd64-all:
	@echo "Building version $(VERSION) for ${GOOS} ${GOARCH}"

	@echo "Building monokit2 for ${GOOS} ${GOARCH}"
	@$(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -o ./bin/monokit2_$(VERSION)_${GOOS}_${GOARCH};

	@for plugin in $(ACTIVE_PLUGINS); do \
		echo "Building plugin $${plugin} for ${GOOS} ${GOARCH}"; \
        cd plugins/$$plugin; \
        $(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -tags $$plugin,${GOOS} -o ../bin/$${plugin}_$(VERSION)_${GOOS}_${GOARCH}; \
        cd ../..; \
    done

linux-arm64-all: GOOS=linux
linux-arm64-all: GOARCH=arm64
linux-arm64-all: TARGET=GOOS=$(GOOS) GOARCH=$(GOARCH)
linux-arm64-all:
	@echo "Building version $(VERSION) for ${GOOS} ${GOARCH}"

	@echo "Building monokit2 for ${GOOS} ${GOARCH}"
	@$(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -o ./bin/monokit2_$(VERSION)_${GOOS}_${GOARCH};

	@for plugin in $(ACTIVE_PLUGINS); do \
		echo "Building plugin $${plugin} for ${GOOS} ${GOARCH}"; \
        cd plugins/$$plugin; \
        $(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -tags $$plugin,${GOOS} -o ../bin/$${plugin}_$(VERSION)_${GOOS}_${GOARCH}; \
        cd ../..; \
    done

windows-amd64-all: GOOS=windows
windows-amd64-all: GOARCH=amd64
windows-amd64-all: TARGET=GOOS=$(GOOS) GOARCH=$(GOARCH)
windows-amd64-all:
	@echo "Building version $(VERSION) for ${GOOS} ${GOARCH}"

	@echo "Building monokit2 for ${GOOS} ${GOARCH}"
	@$(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -o ./bin/monokit2_$(VERSION)_${GOOS}_${GOARCH}.exe;

	@for plugin in $(ACTIVE_PLUGINS); do \
		echo "Building plugin $${plugin} for ${GOOS} ${GOARCH}"; \
        cd plugins/$$plugin; \
        $(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -tags $$plugin,${GOOS} -o ../bin/$${plugin}_$(VERSION)_${GOOS}_${GOARCH}.exe; \
        cd ../..; \
    done

freebsd-amd64-all: GOOS=freebsd
freebsd-amd64-all: GOARCH=amd64
freebsd-amd64-all: TARGET=GOOS=$(GOOS) GOARCH=$(GOARCH)
freebsd-amd64-all:
	@echo "Building version $(VERSION) for ${GOOS} ${GOARCH}"

	@echo "Building monokit2 for ${GOOS} ${GOARCH}"
	@$(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -o ./bin/monokit2_$(VERSION)_${GOOS}_${GOARCH};

	@for plugin in $(ACTIVE_PLUGINS); do \
		echo "Building plugin $${plugin} for ${GOOS} ${GOARCH}"; \
        cd plugins/$$plugin; \
        $(TARGET) go build -ldflags "-X 'main.version=$(VERSION)'" -tags $$plugin,${GOOS} -o ../bin/$${plugin}_$(VERSION)_${GOOS}_${GOARCH}; \
        cd ../..; \
    done

build:
ifeq ($(strip $(PLUGIN)),)
	@echo "No plugin specified. Building main application only."
	@rm -f ./bin/monokit2
	@go build -ldflags "-X main.version=$(VERSION)" -o ./bin/monokit2 ./main.go
else
	@echo "Building $(PLUGIN) with version $(VERSION)"
	@rm -f plugins/bin/$(PLUGIN)
	@cd plugins/$(PLUGIN) && \
	    go build -ldflags "-X main.version=$(VERSION)" -tags $(PLUGIN) -o ../bin/$(PLUGIN)
endif

run: build
ifeq ($(strip $(PLUGIN)),)
	@echo "No plugin specified. Running main application only."
	@./bin/monokit2 $(ARGS)
else
	@echo "Running $(PLUGIN) with version $(VERSION)"
	@./plugins/bin/$(PLUGIN) $(ARGS)
endif

send: build
ifeq ($(strip $(HOST)),)
	@echo "Error: HOST variable is not set. Usage: make send HOST=user@host [PLUGIN=pluginName]"
	@exit 1
endif
	# Host variable is set, make sure plugins directory exists
	@ssh $(HOST) "mkdir -p /var/lib/monokit2/plugins"
ifeq ($(strip $(PLUGIN)),)
	@echo "No plugin specified. Sending main application only."
	@scp ./bin/monokit2 $(HOST):/usr/local/bin/
else
	@echo "Sending $(PLUGIN) to $(HOST)"
	@scp ./plugins/bin/$(PLUGIN) $(HOST):/var/lib/monokit2/plugins/
endif

test-must-run-on-docker:
	@echo "Running tests..."
	@if [ ! -f ./bin/monokit2 ]; then \
        make build; \
    fi
	@./bin/monokit2 reset --force
	@./bin/monokit2
	@TEST=true go test

	@for plugin in $(ACTIVE_PLUGINS); do \
	    cd plugins/$$plugin; \
		TEST=true go test -tags $$plugin; \
		cd ../..; \
	done;

test:
	@docker build -t tests . && docker run --rm tests

test-get-artifacts:
	@docker build -t tests .
	@docker run --rm -v $(realpath ./logs/test):/artifacts -e HOST_UID=$(shell id -u) -e HOST_GID=$(shell id -g) tests

clean:
	@echo "Cleaning ./bin"
	@if [ -d ./bin ]; then \
		for f in ./bin/*; do \
			[ -e "$$f" ] || continue; \
			echo "  removing $$f"; \
			rm -f "$$f"; \
		done; \
	fi

	@echo "Cleaning plugins/bin"
	@if [ -d plugins/bin ]; then \
		for f in plugins/bin/*; do \
			[ -e "$$f" ] || continue; \
			echo "  removing $$f"; \
			rm -f "$$f"; \
		done; \
	fi
