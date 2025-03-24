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

# Build stage
FROM golang:1.24.1-alpine3.21@sha256:43c094ad24b6ac0546c62193baeb3e6e49ce14d3250845d166c77c25f64b0386 AS builder

# Add ca-certificates, timezone data
RUN apk --no-cache add --update ca-certificates tzdata

# Create non-root user
RUN addgroup --gid 65532 golicenser && \
    adduser  --disabled-password --gecos "" \
    --no-create-home --shell="/sbin/nologin" \
    -G golicenser --uid 65532 golicenser

WORKDIR /build/golicenser
COPY . .

RUN go mod download
RUN go mod verify

RUN GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) CGO_ENABLED=0 GOGC=off \
    go build -trimpath -ldflags "-s -w" -o /build/golicenser/bin/golicenser ./cmd/golicenser

# Run stage
FROM scratch

# Build metadata
ARG VERSION
ARG VCS_REF
ARG BUILD_DATE

LABEL maintainer="Joshua Sing <joshua@joshuasing.dev>"
LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.authors="Joshua Sing <joshua@joshuasing.dev>" \
      org.opencontainers.image.url="https://github.com/joshuasing/golicenser" \
      org.opencontainers.image.source="https://github.com/joshuasing/golicenser" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Joshua Sing <joshua@joshuasing.dev>" \
      org.opencontainers.image.title="golicenser" \
      org.opencontainers.image.description="A fast, easy-to-use, license header linter for Go"

# Copy files
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/golicenser/bin/golicenser /usr/local/bin/golicenser

USER golicenser:golicenser
ENTRYPOINT ["/usr/local/bin/golicenser"]
