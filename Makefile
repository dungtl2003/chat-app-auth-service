OUT_DIR = ./bin
OUT_FILE = $(OUT_DIR)/main
SRC_FILES = ./cmd/server/main.go $(wildcard ./internal/**/*.go)

.PHONY: run
run: $(OUT_FILE)
ifdef ENV
	echo "Running application in $(ENV) environment"
	ENV_FILE=.env.$(ENV) ./scripts/run.sh $(OUT_FILE)
else
	echo "Running application in default environment"
	./scripts/run.sh $(OUT_FILE)
endif

.PHONY: build
build: $(OUT_FILE)

$(OUT_FILE): $(SRC_FILES)
	@echo "Building application"
	@mkdir -p $(OUT_DIR)
	go build -o $(OUT_FILE) $<

.PHONY: clean
clean:
	echo "Cleaning up"
	rm -rf $(OUT_DIR)
