include .env
export 

export PROJECT_ROOT=$(shell pwd)

.PHONY: help test 
.DEFAULT_GOAL := help

# help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

settings-cleanup: ## Remove local settings after confirmation
	@read -p "Do you want to clean up your settings? You may lose your data. [y/N]: " ans; \
	if [ "$$ans" == "y" ]; then \
		sudo rm -rf ${PROJECT_ROOT}/out/settings && \
		echo "Settings cleaned up successfuly"; \
	else \
		echo "Cleaning of settings canceled"; \
	fi

logs-cleanup: ## Remove local logs after confirmation
	@read -p "Do you want to clean up your logs? You may lose your data. [y/N]: " ans; \
	if [ "$$ans" == "y" ]; then \
		sudo rm -rf ${PROJECT_ROOT}/out/logs && \
		echo "Logs cleaned up successfuly"; \
	else \
		echo "Cleaning of logs canceled"; \
	fi

run: ## Run the terminal UI locally
	@mkdir -p ${PROJECT_ROOT}/out/logs ${PROJECT_ROOT}/out/settings 2>/dev/null || \
		(echo "Cannot create local state directories. Run: sudo chown -R $$(id -u):$$(id -g) ${PROJECT_ROOT}/out"; exit 1)
	@test -w ${PROJECT_ROOT}/out/logs -a -w ${PROJECT_ROOT}/out/settings || \
		(echo "Local state directories are not writable. Run: sudo chown -R $$(id -u):$$(id -g) ${PROJECT_ROOT}/out"; exit 1)
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export SETTINGS_FILE=${PROJECT_ROOT}/out/settings/settings.json && \
	go run ${PROJECT_ROOT}/cmd/ankiwords/main.go 

# other
test: ## Run Go tests
	@go test ${PROJECT_ROOT}/...
