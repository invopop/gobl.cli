FROM golang:1.18.2-alpine AS go-build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o gobl.wasm ./wasm

FROM nginx:1.21.6
WORKDIR /usr/share/nginx/html
COPY --from=go-build /usr/local/go/misc/wasm/wasm_exec.js ./
COPY --from=go-build /src/gobl.wasm ./
COPY wasm/*.html ./
COPY wasm/*.js ./
