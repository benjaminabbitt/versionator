# Versionator Builder Base Image
# Build once, use in all language test containers

FROM golang:1.23-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /versionator .

# Create a minimal runtime stage with just the binary
FROM scratch AS binary
COPY --from=builder /versionator /versionator
