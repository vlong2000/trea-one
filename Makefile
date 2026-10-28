GO_BIN_PATH := bin
NRF_SRC := nfs/nrf

.PHONY: nrf nfs test-nrf

nfs: nrf

nrf:
	mkdir -p $(GO_BIN_PATH)
	CGO_ENABLED=0 go build -C $(NRF_SRC) -o ../../$(GO_BIN_PATH)/nrf ./cmd

test-nrf:
	go test -C $(NRF_SRC) ./...
