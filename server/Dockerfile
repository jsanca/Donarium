# Stage 1: build
FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/donarium ./cmd/donarium/

# Stage 2: runtime
FROM alpine:3.22

RUN addgroup -S donarium && adduser -S donarium -G donarium

WORKDIR /app

COPY --from=build /bin/donarium .

RUN chown -R donarium:donarium /app

USER donarium

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${HTTP_PORT:-8080}/health/live || exit 1

ENTRYPOINT ["./donarium"]
