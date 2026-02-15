# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/voidvpn/voidvpn/pkg/version.Version=${VERSION} \
    -X github.com/voidvpn/voidvpn/pkg/version.Commit=${COMMIT} \
    -X github.com/voidvpn/voidvpn/pkg/version.BuildDate=${BUILD_DATE}" \
    -o /app/voidvpn ./cmd/voidvpn

# Runtime stage
FROM alpine:3.19 AS runtime
RUN apk --no-cache add ca-certificates iptables iproute2

WORKDIR /app
COPY --from=build /app/voidvpn .

# Run as non-root user
RUN adduser -D -u 1000 voidvpn
USER voidvpn

ENTRYPOINT ["./voidvpn"]
