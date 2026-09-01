# Multi-stage: build the Go binary, then run it in a slim runtime.
#
# Build:   docker build -t ghcr.io/ramackersjp/taxicheck .
# Run:     docker run -it --rm ghcr.io/ramackersjp/taxicheck
#
# The image is tagged with the release version (e.g. 2.0.1) and `latest` on
# release by .github/workflows/docker.yml.

FROM golang:1.27 AS builder
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -ldflags "-s -w" -o /out/taxiprijs ./cmd/taxiprijs

FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/taxiprijs /usr/local/bin/taxiprijs
ENV HOME=/root
WORKDIR /root
# Network to reach OSRM/PDOK/Nominatim; config is stored under ~/.taxiprijs.
CMD ["taxiprijs"]