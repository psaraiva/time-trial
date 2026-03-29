# time-trial

```
╔═════════════════════════════════════════════════╗
║                                                 ║
║   _____ _                _____     _       _    ║
║  |_   _(_)_ __ ___   ___|_   _| __(_) __ _| |   ║
║    | | | | '_ ` _ \ / _ \ | || '__| |/ _` | |   ║
║    | | | | | | | | |  __/ | || |  | | (_| | |   ║
║    |_| |_|_| |_| |_|\___| |_||_|  |_|\__,_|_|   ║
║                                                 ║
╚═════════════════════════════════════════════════╝
```

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Language: Português](https://img.shields.io/badge/Language-Portugu%C3%AAs-green?style=flat-square)](./README.pt-BR.md)

[![CI](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml/badge.svg)](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/time-trial)](https://goreportcard.com/report/github.com/psaraiva/time-trial)
[![codecov](https://codecov.io/gh/psaraiva/time-trial/branch/main/graph/badge.svg)](https://codecov.io/gh/psaraiva/time-trial)

A minimal Go application that exposes endpoints to simulate dependent-service behavior — forcing specific HTTP status codes, delays, and dynamic response bodies in dependent services or tests.

## Overview

The application exposes HTTP routes built with [Fiber](https://gofiber.io/). You can set a forced status code with a delay range (`POST /time-trial`), load an ordered plan of configurations (`POST /plan`), configure a response-body schema (`POST /param-resp`), and trigger executions that follow those configurations (`GET /sabotage`, `GET /plan/sabotage`). State is held in memory with lock-free atomics and mutex, safe for concurrent use.

When a `param-resp` configuration is active and an execution returns **200**, the response body is generated dynamically according to the configured schema instead of the standard sabotage metadata.

Every response is wrapped in a standard envelope:

```json
{
  "data":      { ... },
  "duration":  42,
  "timestamp": "2026-03-19T10:00:00Z"
}
```

## Requirements

- Go 1.25+
- `jq` (optional, for formatted endpoint output)

## Getting Started

```bash
make deps
make run
```

The server listens on port `7777` by default. Override with the `TIME_TRIAL_API_PORT` environment variable:

```bash
TIME_TRIAL_API_PORT=8080 make run
```

## Swagger UI

Start the server, then open:

```
http://localhost:7777/swag/index.html
```

The interactive UI documents all available routes, request bodies, and response schemas.

To regenerate the Swagger spec after changing annotations:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swag
```

## Routes

| Method | Path                 | Description                                                      |
|--------|----------------------|------------------------------------------------------------------|
| POST   | `/time-trial`        | Set or reset the forced status code and delay range              |
| GET    | `/time-trial/config` | Return the current sabotage configuration                        |
| GET    | `/sabotage`          | Execute a request using the active sabotage configuration        |
| POST   | `/plan`              | Set or clear an ordered plan of sabotage configurations          |
| GET    | `/plan/sabotage`     | Execute the next step of the active plan                         |
| GET    | `/plan/config`       | Return the full active plan                                      |
| POST   | `/param-resp`        | Set or clear the response-body generation configuration          |
| GET    | `/param-resp/config` | Return the current param-resp configuration                      |

---

### POST /time-trial

Sets the active sabotage configuration. With no body, resets to random behavior.

**Body:**
```json
{
  "code":     500,
  "delayMin": 100,
  "delayMax": 900
}
```

| Field      | Accepted values                     | Description                       |
|------------|-------------------------------------|-----------------------------------|
| `code`     | `0`, `200`, `400`, `500`            | Forced status code (`0` = reset)  |
| `delayMin` | `1`–`60000` ms, or `0` to disable   | Minimum delay in milliseconds     |
| `delayMax` | `>= delayMin`, max `60000` ms       | Maximum delay in milliseconds     |

**No body** → resets sabotage (`code=0`, delays zeroed).

---

### GET /sabotage

Executes a simulated request using the active sabotage configuration. If no sabotage is active (`code=0`), responds with a random code (200, 400 or 500).

When the resulting code is **200** and a `param-resp` configuration is active, the response body is generated dynamically — see [POST /param-resp](#post-param-resp).

---

### GET /time-trial/config

Returns the current sabotage configuration.

**Response:**
```json
{
  "sabotaged": true,
  "code":      500,
  "delayMin":  100,
  "delayMax":  900
}
```

---

### POST /plan

Loads an ordered plan of sabotage configs. Each item is a full sabotage configuration (State). Steps are consumed in order on each `GET /plan` call.

**Body:**
```json
{
  "plan": [
    { "code": 500, "delayMin": 100, "delayMax": 500 },
    { "code": 200, "delayMin": 50,  "delayMax": 200 }
  ]
}
```

**No body** → clears and cancels the active plan from memory.

---

### GET /plan/sabotage

Executes the next step of the active plan. Returns `404` if no plan is loaded, all steps have been consumed, or the plan was interrupted.

When the resulting code is **200** and a `param-resp` configuration is active, the response body is generated dynamically — see [POST /param-resp](#post-param-resp).

---

### GET /plan/config

Returns the full active plan regardless of how many steps have been consumed. Returns `404` if no plan is loaded.

**Response:**
```json
{
  "active": true,
  "steps": [
    { "code": 500, "delayMin": 100, "delayMax": 500 },
    { "code": 200, "delayMin": 50,  "delayMax": 200 }
  ]
}
```

---

### POST /param-resp

Configures the schema used to generate response bodies when a `200` is returned by `/sabotage` or `/plan`. With no body, clears the active configuration.

Supported property types: `string`, `int`, `float`.

**Body:**
```json
{
  "statusCode": 200,
  "item": {
    "isColection": true,
    "quantity": 5,
    "properties": [
      {
        "name": "productName",
        "type": "string",
        "isRequired": true,
        "maxLength": 10,
        "minLength": 3,
        "propertyString": {
          "chars": "abcdefghijklmnopqrstuvxzABCDEFGHIJKLMNOPQRSTUVXZ"
        }
      },
      {
        "name": "value",
        "type": "float",
        "isRequired": true,
        "maxLength": 10000,
        "minLength": 0,
        "propertyFloat": {
          "floatPrecision": 2,
          "isAcceptNegativeValue": false
        }
      },
      {
        "name": "version",
        "type": "int",
        "isRequired": true,
        "maxLength": 99,
        "minLength": 0,
        "propertyInt": {
          "isAcceptNegativeValue": false
        }
      }
    ]
  }
}
```

**Top-level fields:**

| Field        | Description                                  |
|--------------|----------------------------------------------|
| `statusCode` | Must be `200` (only supported value)         |
| `item`       | Describes the item or collection to generate |

**`item` fields:**

| Field          | Description                                                                     |
|----------------|---------------------------------------------------------------------------------|
| `isColection`  | `true` → returns a JSON array `[]`; `false` → returns a single object `{}`     |
| `quantity`     | Number of items to generate (ignored when `isColection` is `false`)             |
| `properties`   | 1..N property definitions                                                       |

**Common property fields:**

| Field        | Description                                                                         |
|--------------|-------------------------------------------------------------------------------------|
| `name`       | JSON key name (letters, digits, `_`, `-` only)                                      |
| `type`       | `"string"`, `"int"`, or `"float"`                                                   |
| `isRequired` | `false` → field is included in the response with value `null`                       |
| `maxLength`  | For `string`: max character length. For `int`/`float`: maximum value               |
| `minLength`  | For `string`: min character length. For `int`/`float`: minimum value               |

**`propertyString` fields:**

| Field   | Description                                        |
|---------|----------------------------------------------------|
| `chars` | Character set for generation (letters A–Z, a–z only) |

**`propertyInt` fields:**

| Field                   | Description                         |
|-------------------------|-------------------------------------|
| `isAcceptNegativeValue`  | Allow negative values in generation |

**`propertyFloat` fields:**

| Field                   | Description                                  |
|-------------------------|----------------------------------------------|
| `floatPrecision`        | Number of decimal places                     |
| `isAcceptNegativeValue`  | Allow negative values in generation          |

**No body** → clears the active configuration.

---

### GET /param-resp/config

Returns the current param-resp configuration exactly as it was submitted. Returns `404` if no configuration is active.

---

## Makefile Targets

```
make deps      Download and tidy Go dependencies
make build     Compile binary to bin/server
make run       Run server locally
make swag      Generate Swagger docs (requires swag CLI)
make vet       Run go vet
make lint      Run golangci-lint
make ci        Run vet, lint and tests (mirrors CI pipeline)
make test      Run unit tests with race detector
make coverage  Generate coverage report
make help      List all targets with descriptions
```

---

## License

Copyright (c) 2026 time-trial contributors.

This project is licensed under the **Apache License 2.0**.

You are free to use, modify, and distribute this software under the terms of the license.

See the [`LICENSE`](LICENSE) file for full terms or visit
[apache.org/licenses/LICENSE-2.0](https://www.apache.org/licenses/LICENSE-2.0).
