
# Build Nuxt frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --shamefully-hoist
COPY frontend .
RUN pnpm build

# Build PocketBase binary
FROM golang:1.25-alpine AS builder
ARG BUILD_TIME
ARG COMMIT
ARG VERSION
RUN apk add --no-cache git

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-builder /app/.output/public ./pb_public
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME -X main.version=$VERSION" \
    -o pocketbase \
    ./app/api/

# Production
FROM alpine:latest

ENV HBOX_MODE=production
ENV HBOX_WEB_HOST=0.0.0.0
ENV HBOX_STORAGE_DATA=/data/
ENV HBOX_STORAGE_POCKETBASE_DIR=/data/pb_data
ENV HBOX_PUBLIC_DIR=/pb/pb_public

RUN apk --no-cache add ca-certificates && mkdir -p /pb
COPY --from=builder /build/pocketbase /pb/pocketbase
COPY --from=builder /build/pb_public /pb/pb_public/
RUN chmod +x /pb/pocketbase

LABEL Name=homebox Version=0.0.1
LABEL org.opencontainers.image.source="https://github.com/compdani/homebox"
EXPOSE 7745
WORKDIR /pb

ENTRYPOINT ["/pb/pocketbase"]
