FROM golang:1.21 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go vet
RUN go test
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o app ./cmd/server

FROM gcr.io/distroless/static-debian11

COPY --from=builder app /
EXPOSE 8080
CMD ["/app"]
