ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
OUT_DIR = ./bin
OUT_FILE = $(OUT_DIR)/main
SRC_FILES = ./cmd/server/main.go $(shell find ./internal/ -name '*.go')

.PHONY: test
test: export TEST_OUT = $(ROOT_DIR)/reports/results
test: export DB_LOG = $(ROOT_DIR)/reports/db.log
test: export USER_SERVICE_LOG = $(ROOT_DIR)/reports/user_service.log
test: export SNOWFLAKE_SERVICE_LOG = $(ROOT_DIR)/reports/snowflake_service.log
test: build certs
	@rm -rf reports
	@mkdir reports
	@echo "Running tests"
ifdef JSON
	TEST_OUT=$(TEST_OUT) ./scripts/test_local.sh go run ./cmd/test/run_tests.go -json
	./scripts/read_test_stats.sh $(TEST_OUT)
else
	./scripts/test_local.sh go run ./cmd/test/run_tests.go
endif

.PHONY: run
run: build
ifdef ENV
	@echo "Running application in $(ENV) environment"
	ENV_FILE=.env.$(ENV) ./scripts/run.sh $(OUT_FILE)
else
	@echo "Running application in default environment"
	./scripts/run.sh $(OUT_FILE)
endif

.PHONY: build
build: $(OUT_FILE)

$(OUT_FILE): $(SRC_FILES)
	@echo "$(SRC_FILES)"
	@echo "Building application"
	@mkdir -p $(OUT_DIR)
	go build -o $(OUT_FILE) $<

.PHONY: clean
clean:
	@echo "Cleaning up"
	rm -rf $(OUT_DIR)

.PHONY: certs
certs:
	@echo "Generating certs"	
	./scripts/gen_certs.sh
