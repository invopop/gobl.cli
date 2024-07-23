# GOBL CLI

<img src="https://github.com/invopop/gobl/blob/main/gobl_logo_black_rgb.svg?raw=true" width="181" height="219" alt="GOBL Logo">

Go Business Language - Command Line Interface

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2021-2023 [Invopop Ltd.](https://invopop.com).

**DEPRECATED**: The CLI has now been moved back to the main [gobl](https://github.com/invopop/gobl) repository.

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

### Build

Build expects a partial GOBL Envelope or Document, in either YAML or JSON as input. It'll automatically run the Calculate and Validate methods and output JSON data as either an envelope or document, according to the input source.

Example uses:

```sh
# Calculate and validate a YAML invoice
gobl build ./samples/invoice-es.yaml

# Output using indented formatting
gobl build -i ./samples/customer.yaml

# Set the supplier from an external file
gobl build -i ./samples/invoice-es.yaml \
    --set-file customer=./samples/customer.yaml

# Set arbitrary values from the command line. Inputs are parsed as YAML.
gobl build -i ./samples/invoice-es.yaml \
    --set meta.bar="a long string" \
    --set series="TESTING"

# Set the top-level object:
gobl build -i ./samples/invoice-es.yaml \
    --set-file .=./samples/envelope-invoice-es.yaml

# Insert a document into an envelope
gobl build -i --envelop ./samples/invoice-es.yaml
```

### Correct

The GOBL CLI makes it easy to use the library and tax regime specific functionality that create a corrective document that reverts or amends a previous document. This is most useful for invoices and issuing refunds for example.

```sh
# Correct an invoice with a credit note (this will error for ES invoice!)
gobl correct -i ./samples/invoice-es.yaml --credit

# Specify tax regime specific details
gobl correct -i -d '{"credit":true,"changes":["line"],"method":"complete"}' \
    ./samples/invoice-es.yaml
```

### Sign

GOBL encourages users to sign data embedded into envelopes using digital signatures. To get started, you'll need to have a JSON Web Key. Use the following commands to generate one:

```sh
# Generate a JSON Web Key and store in ~/.gobl/id_es256.jwk
gobl keygen

# Generate and output a JWK into a new file
gobl keygen ./samples/key.jwk
```

Use the key to sign documents:

```sh
# Add a signature to the envelope using our personal key
gobl sign -i ./samples/envelope-invoice-es.yaml

# Add a signature using a specific key
gobl sign -i --key ./samples/key.jwk ./samples/envelope-invoice-es.yaml
```

It is only possible to sign non-draft envelopes, so the CLI will automatically remove this flag during the signing process. This implies that the document must be completely valid before signing.
