This folder contains the files necessary to execute the GOBL client library in your web browser. The core library is written in Go, and compiled to WebAssembly.  The file `gobl.js` provides a thin JavaScript wrapper around the compiled WebAssembly running in a web worker.

To execute a simple demo of the GOBL library in your browser, you will need to:

1. [Install Go](https://go.dev/dl/) (1.17 or newer)
2. From this directory, execute the command `./build.sh` which will compile the WASM target and start a bare web server.
3. In your browser, navigate to `http://localhost:9999/`
4. Open the JavaScript console in your browser to see the test output.
