# time-trial

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Language: Português](https://img.shields.io/badge/Language-Portugu%C3%AAs-green?style=flat-square)](./README.pt-BR.md)

[![CI](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml/badge.svg)](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/time-trial?style=flat)](https://goreportcard.com/report/github.com/psaraiva/time-trial?force=true)
[![codecov](https://codecov.io/gh/psaraiva/time-trial/branch/main/graph/badge.svg)](https://codecov.io/gh/psaraiva/time-trial)

A minimal Go application that exposes endpoints to control a sabotage state — forcing specific HTTP status codes and delays in dependent services or tests.

## Overview

The application exposes HTTP routes built with [Fiber](https://gofiber.io/). You can set a forced status code with a delay range (`POST /sabotage`), load an ordered plan of configurations (`POST /plan/sabotage`), and trigger executions that follow those configurations (`GET /sabotage/exec`, `GET /plan/exec`). State is held in memory with lock-free atomics and mutex, safe for concurrent use.

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

## Routes

| Method | Path                 | Description                                                      |
|--------|----------------------|------------------------------------------------------------------|
| POST   | `/sabotage`          | Set or reset the forced status code and delay range              |
| GET    | `/sabotage/exec`     | Execute a request using the active sabotage configuration        |
| GET    | `/sabotage/config`   | Return the current sabotage configuration                        |
| POST   | `/plan/sabotage`     | Load or clear an ordered plan of sabotage configurations         |
| GET    | `/plan/exec`         | Execute the next step of the active plan                         |
| GET    | `/plan/config`       | Return the full active plan                                      |

---

### POST /sabotage

Sets the active sabotage configuration. With no body, resets to random behavior.

**Body:**
```json
{
  "code":     500,
  "delayMin": 100,
  "delayMax": 900
}
```

| Field      | Accepted values                    | Description                       |
|------------|------------------------------------|-----------------------------------|
| `code`     | `0`, `200`, `400`, `500`           | Forced status code (`0` = reset)  |
| `delayMin` | `1`–`60000` ms, or `0` to disable   | Minimum delay in milliseconds     |
| `delayMax` | `>= delayMin`, max `60000` ms       | Maximum delay in milliseconds     |

**No body** → resets sabotage (`code=0`, delays zeroed).

---

### GET /sabotage/exec

Executes a simulated request using the active sabotage configuration. If no sabotage is active (`code=0`), responds with a random code (200, 400 or 500).

---

### GET /sabotage/config

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

### POST /plan/sabotage

Loads an ordered plan of sabotage configs. Each item is a full sabotage configuration (State). Steps are consumed in order on each `GET /plan/exec` call.

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

### GET /plan/exec

Executes the next step of the active plan. Returns `404` if no plan is loaded, all steps have been consumed, or the plan was interrupted.

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

## Makefile Targets

```
make deps      Download and tidy Go dependencies
make build     Compile binary to bin/server
make run       Run server locally
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
