# jdw-go

[![Go Report Card](https://goreportcard.com/badge/github.com/KRoperUK/get_spoons)](https://goreportcard.com/report/github.com/KRoperUK/get_spoons)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/KRoperUK/get_spoons)](https://github.com/KRoperUK/get_spoons)
[![Test](https://github.com/KRoperUK/get_spoons/actions/workflows/test.yml/badge.svg)](https://github.com/KRoperUK/get_spoons/actions/workflows/test.yml)
[![Lint](https://github.com/KRoperUK/get_spoons/actions/workflows/lint.yml/badge.svg)](https://github.com/KRoperUK/get_spoons/actions/workflows/lint.yml)
[![release-please](https://github.com/KRoperUK/get_spoons/actions/workflows/release-please.yaml/badge.svg)](https://github.com/KRoperUK/get_spoons/actions/workflows/release-please.yaml)

A Go client library for the J.D. Wetherspoon (JDW) API, including a companion CLI tool for scraping pub data.

## Features

- **Reusable Library**: Direct access to the JDW API via the `jdw` Go package.
- **REST & GraphQL Support**: Wraps common endpoints for venues, settings, and promotional content.
- **CLI Utility**: `get_spoons` tool for generating CSV datasets.
- **Fast & Authenticated**: Uses identified Bearer tokens and app headers for reliable access.

## Repository Structure

- `jdw/`: The Go library package.
- `cmd/get_spoons/`: Source code for the CLI tool.
- `openapi.yaml`: Unofficial OpenAPI 3.0 specification for the JDW API.

## Installation

```bash
go get github.com/KRoperUK/get_spoons/jdw
```

To install the CLI tool:

```bash
go install github.com/KRoperUK/get_spoons/cmd/get_spoons@latest
```

## CLI Usage

The CLI tool requires a JDW Bearer Token (captured manually, starting with `1|...`).

**Option 1: Environment Variable (Recommended)**

```bash
export JDW_TOKEN="1|..."
get_spoons --output my_pubs.csv
```

**Option 2: Flag**

```bash
get_spoons --token "1|..." --output my_pubs.csv
```

## Library Usage

```go
import "github.com/KRoperUK/get_spoons/jdw"

client := jdw.NewClient("6.7.1", "YOUR_TOKEN", "YOUR_USER_AGENT")
venues, err := client.GetVenues()
```

## Configuration

The library and CLI tool require a JDW Bearer Token for authentication. You can provide this via the `JDW_TOKEN` environment variable or the `--token` CLI flag.

- `JDW_TOKEN`: Your API bearer token.
- `JDW_APP_VERSION`: (Optional) JDW app version (default: `6.7.1`).
- `JDW_USER_AGENT`: (Optional) Custom user agent.

## Testing

- `make test`: Runs unit tests with mocked responses (no API key required).
- `make test-live`: Runs integration tests against the live JDW API.
  - **Requires** `JDW_TOKEN` (Manually captured `Bearer 1|...` token).
  - Automated token generation is not currently supported for the main API endpoints.
  ```bash
  JDW_TOKEN="1|..." make test-live
  ```

## API Documentation

See [openapi.yaml](openapi.yaml) for a full description of the identified endpoints.

## Web Preview

View the latest .csv of data at: [spoons.ink](https://spoons.ink)
