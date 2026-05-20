# Stage 1: Build the web UI with Node.
FROM node:22-alpine AS ui

WORKDIR /ui
COPY ui/package.json ui/package-lock.json* ui/.npmrc* ./
RUN npm install --no-audit --no-fund --loglevel=error
COPY ui ./
ENV NODE_OPTIONS=--max-old-space-size=3072
RUN npm run build

# Stage 2: Build the Go binary with the UI assets embedded.
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Drop whatever dist shipped in the repo and copy in the freshly-built UI.
RUN rm -rf ui/dist
COPY --from=ui /ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /kiwifs .

# Stage 3: Minimal runtime with document export dependencies.
FROM alpine:3.20

# Core system packages.
RUN apk add --no-cache git ca-certificates docker-cli \
    # Pandoc for PDF/HTML export.
    pandoc \
    # Node.js runtime for Marp CLI.
    nodejs npm \
    # Python for MkDocs.
    python3 py3-pip \
    # Chromium for Marp PDF/PPTX export (headless).
    chromium \
    && addgroup -S kiwi && adduser -S kiwi -G kiwi

# Install Marp CLI globally.
RUN npm install -g @marp-team/marp-cli --no-audit --no-fund 2>/dev/null || true

# Install MkDocs with Material theme.
RUN pip3 install --break-system-packages --no-cache-dir \
    mkdocs mkdocs-material 2>/dev/null || true

# Set Chromium path for Marp's headless PDF/PPTX export.
ENV CHROME_PATH=/usr/bin/chromium-browser
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium-browser

COPY --from=builder /kiwifs /usr/local/bin/kiwifs

RUN mkdir -p /data && chown kiwi:kiwi /data

USER kiwi

EXPOSE 3333

VOLUME ["/data"]

ENTRYPOINT ["kiwifs"]
CMD ["serve", "--root", "/data", "--port", "3333", "--host", "0.0.0.0", "--search", "sqlite"]
