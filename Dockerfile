FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o pr-reviewer ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/pr-reviewer ./pr-reviewer
EXPOSE 8080
ENTRYPOINT ["/app/pr-reviewer"]
