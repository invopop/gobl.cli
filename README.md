# GOBL CLI

<img src="https://github.com/invopop/gobl/blob/main/gobl_logo_black_rgb.svg?raw=true" width="181" height="219" alt="GOBL Logo">

Go Business Language - Command Line Interface

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2021,2022 [Invopop Ltd.](https://invopop.com).

## Usage

This repo contains a `gobl` CLI tool which can be used to manipulate GOBL documents from the command line or shell scripts.

Build with:

```sh
mage -v build
```

Install with:

```sh
mage -v install
```

### `gobl build`

Build a complete GOBL document from one or more input sources. Example uses:

```sh
# Finalize a complete invoice
gobl build invoice.yaml

# Set the supplier from an external file
gobl build invoice.yaml \
    --set-file doc.supplier=supplier.yaml

# Set arbitrary values from the command line. Inputs are parsed as YAML.
gobl build invoice.yaml \
    --set doc.foo.bar="a long string" \
    --set doc.foo.baz=1234

# Set an explicit string value (to avoid interpetation as a boolean or number)
gobl build invoice.yaml \
    --set-string doc.foo.baz=1234 \
    --set-string doc.foo.quz=true

# Set the top-level object:
gobl build invoice.yaml \
    --set-file .=envelope.yaml
```
