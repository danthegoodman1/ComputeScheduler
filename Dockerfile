FROM golang:1.21.1-bullseye as build

WORKDIR /app

# START GIT PRIVATE SECTION - delete if not using private packages
ARG GIT_INSTEAD_OF=ssh://git@github.com/
ARG GO_ARGS=""

# Install docker
RUN curl -fsSL https://get.docker.com | sh

# Install overmind
RUN go install github.com/DarthSim/overmind/v2@latest && mv $(go env GOPATH)/bin/overmind /

# Install serf and tmux
RUN apt update && apt install unzip tmux -y
RUN wget https://releases.hashicorp.com/serf/0.8.2/serf_0.8.2_linux_amd64.zip -O serf.zip \
    && unzip serf.zip && mv serf / && chmod a+x /serf

COPY go.* /app/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=ssh \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build $GO_ARGS -o /app/outbin

ENV OVERMIND_NO_PORT=1
ENV OVERMIND_TIMEOUT=30
ENV OVERMIND_PROCFILE=/app/procfile
#ENV OVERMIND_CAN_DIE=serf_setup
CMD overmind start