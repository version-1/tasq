FROM golang:1.22-bookworm

ARG NODE_MAJOR=22

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      bubblewrap \
      ca-certificates \
      curl \
      git \
      gh \
      gnupg \
      make \
      procps \
      sqlite3 \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_${NODE_MAJOR}.x nodistro main" > /etc/apt/sources.list.d/nodesource.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && npm install -g @openai/codex \
    && useradd --create-home --shell /bin/bash codex \
    && mkdir -p /workspace /home/codex/.codex /home/codex/.cache/go-build /go/pkg/mod \
    && chown -R codex:codex /workspace /home/codex /go/pkg/mod \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV TQ_HOME=/workspace/.tasq
ENV CODEX_HOME=/home/codex/.codex
ENV GOCACHE=/home/codex/.cache/go-build
ENV PATH=/usr/local/go/bin:$PATH

WORKDIR /workspace
