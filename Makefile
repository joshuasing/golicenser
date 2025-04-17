# Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

.PHONY: all
all: lint build test run

.PHONY: deps
deps:
	go mod download
	go mod verify

.PHONY: lint
lint: tidy
	golangci-lint run --fix ./...

.PHONY: lint-deps
lint-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	go build ./...
	go build -trimpath -ldflags '-s -w' -o ./bin/golicenser ./cmd/golicenser

.PHONY: run
run: build
	./bin/golicenser -year-mode=git-range -tmpl=MIT -author='Joshua Sing <joshua@joshuasing.dev>' -fix ./...

.PHONY: test
test:
	go test -v -coverprofile=.cover ./...

.PHONY: cover
cover:
	go tool cover -html .cover

.PHONY: test-race
test-race:
	go test -race ./...
