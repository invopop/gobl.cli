FROM golang:1.18.2-alpine AS go-build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o gobl ./cmd/gobl

FROM alpine:3.15
COPY --from=go-build /src/gobl ./
EXPOSE 80/tcp
CMD ["./gobl","serve"]
