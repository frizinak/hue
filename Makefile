dist/hue: $(shell find -type f -name '*.go')
	go build -o $@ ./cmd/hue

.PHONY: install
install:
	go install ./cmd/hue

