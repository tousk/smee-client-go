<h2 align="center">smee-client-go</h2>
<p align="center">A Golang-based client for smee.io, a service that delivers webhooks to your local development environment.</p>

## Installation

```
$ make
```

## Usage

### CLI

The `smee` command will forward webhooks from smee.io to your local development environment. It also supports github's authenticated header.

Run `smee --help` for usage.

```
$ ./smee --help
Usage:
  smee [OPTIONS]

Application Options:
  -v, --version  output the version number
  -u, --url=     URL of the webhook proxy service. Required.
  -t, --target=  Full URL (including protocol and path) of the target service the events will forwarded to. Required.
  -s, --secret=  Secret to be used for HMAC-SHA1 secure hash calculation

Help Options:
  -h, --help     Show this help message

```

