ACTIVE_PLUGINS := osHealth
VERSION ?= devel

# Example usage:
# make all VERSION=1.0.0
# make build PLUGIN=osHealth VERSION=1.0.0
# make run PLUGIN=osHealth ARGS="-v" VERSION=1.0.0

.PHONY: clean all

all:
	@echo "Building version $(VERSION)"

	@if [ -f ./bin/monokit2 ]; then \
		rm ./bin/monokit2; \
	fi

	@echo "Building monokit2"
	@go build -ldflags "-X 'main.version=$(VERSION)'" -o ./bin/monokit2 ./main.go;

	@for plugin in $(ACTIVE_PLUGINS); do \
		echo "Building plugin $$plugin"; \
        if [ -f plugins/bin/$$plugin ]; then \
            rm plugins/bin/$$plugin; \
        fi; \
        cd plugins/$$plugin; \
        go build -ldflags "-X 'main.version=$(VERSION)'" -tags $$plugin -o ../bin/; \
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

test:
	@echo "Running tests..."
	@./bin/monokit2 reset --force
	@./bin/monokit2
	@TEST=true go test

	@for plugin in $(ACTIVE_PLUGINS); do \
	    cd plugins/$$plugin; \
		TEST=true go test -tags $$plugin; \
		cd ../..; \
	done;

test-docker:
	@docker build -t tests . && docker run --rm tests

clean:
	@if [ -f ./bin/monokit2 ]; then \
	    rm ./bin/monokit2; \
	fi

	@for plugin in $(ACTIVE_PLUGINS); do \
		if [ -f plugins/bin/$$plugin ]; then \
			rm plugins/bin/$$plugin; \
		fi; \
	done
