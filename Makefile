BINARY = heimdallr-sense
BUILD_DIR = build
CMD = ./cmd/vad

.PHONY: all clean tidy

all: $(BUILD_DIR)/$(BINARY)-linux-amd64 \
     $(BUILD_DIR)/$(BINARY)-linux-arm64 \
     $(BUILD_DIR)/$(BINARY)-linux-armv7

tidy:
	go mod tidy

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY)-linux-amd64: tidy $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags="-s -w" -o $@ $(CMD)

$(BUILD_DIR)/$(BINARY)-linux-arm64: tidy $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -trimpath -ldflags="-s -w" -o $@ $(CMD)

$(BUILD_DIR)/$(BINARY)-linux-armv7: tidy $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
		go build -trimpath -ldflags="-s -w" -o $@ $(CMD)

clean:
	rm -rf $(BUILD_DIR)
