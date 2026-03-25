# time-trial

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Language: English](https://img.shields.io/badge/Idioma-English-blue?style=flat-square)](./README.md)

[![CI](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml/badge.svg)](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/time-trial)](https://goreportcard.com/report/github.com/psaraiva/time-trial)
[![codecov](https://codecov.io/gh/psaraiva/time-trial/branch/main/graph/badge.svg)](https://codecov.io/gh/psaraiva/time-trial)

Aplicação Go minimalista que expõe endpoints para controlar um estado de sabotagem — forçando status codes HTTP específicos e delays em serviços dependentes ou testes.

## Visão Geral

A aplicação expõe rotas HTTP construídas com [Fiber](https://gofiber.io/). É possível definir um status code forçado com range de delay (`POST /sabotage`), carregar um plano ordenado de configurações (`POST /plan/sabotage`), e disparar execuções que respeitam essas configurações (`GET /sabotage/exec`, `GET /plan/exec`). O estado é mantido em memória com atomics lock-free e mutex, sendo seguro para uso concorrente.

Toda resposta é encapsulada em um envelope padrão:

```json
{
  "data":      { ... },
  "duration":  42,
  "timestamp": "2026-03-19T10:00:00Z"
}
```

## Requisitos

- Go 1.25+
- `jq` (opcional, para saída formatada nos endpoints)

## Como Começar

```bash
make deps
make run
```

O servidor escuta na porta `7777` por padrão. Sobrescreva com a variável de ambiente `TIME_TRIAL_API_PORT`:

```bash
TIME_TRIAL_API_PORT=8080 make run
```

## Rotas

| Método | Caminho              | Descrição                                                         |
|--------|----------------------|-------------------------------------------------------------------|
| POST   | `/sabotage`          | Define ou reseta o status code forçado e range de delay           |
| GET    | `/sabotage/exec`     | Executa uma requisição com a configuração ativa de sabotagem      |
| GET    | `/sabotage/config`   | Retorna a configuração atual de sabotagem                         |
| POST   | `/plan/sabotage`     | Carrega ou limpa um plano ordenado de configurações de sabotagem  |
| GET    | `/plan/exec`         | Executa o próximo passo do plano ativo                            |
| GET    | `/plan/config`       | Retorna o plano completo ativo                                    |

---

### POST /sabotage

Define a configuração ativa de sabotagem. Sem corpo, reseta para comportamento aleatório.

**Body:**
```json
{
  "code":     500,
  "delayMin": 100,
  "delayMax": 900
}
```

| Campo      | Valores aceitos                        | Descrição                        |
|------------|----------------------------------------|----------------------------------|
| `code`     | `0`, `200`, `400`, `500`               | Status code forçado (`0` = reset)|
| `delayMin` | `1`–`60000` ms, ou `0` para desativar   | Delay mínimo em milissegundos    |
| `delayMax` | `>= delayMin`, máximo `60000` ms        | Delay máximo em milissegundos    |

**Sem corpo** → reseta sabotage (`code=0`, delays zerados).

---

### GET /sabotage/exec

Executa uma requisição simulada usando a configuração ativa de sabotagem. Se nenhuma sabotagem estiver ativa (`code=0`), responde com um código aleatório (200, 400 ou 500).

---

### GET /sabotage/config

Retorna a configuração atual de sabotagem.

**Resposta:**
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

Carrega um plano ordenado de sabotagens. Cada item é uma configuração completa de sabotagem (State). Os passos são consumidos em ordem a cada chamada de `GET /plan/exec`.

**Body:**
```json
{
  "plan": [
    { "code": 500, "delayMin": 100, "delayMax": 500 },
    { "code": 200, "delayMin": 50,  "delayMax": 200 }
  ]
}
```

**Sem corpo** → limpa e cancela o plano ativo da memória.

---

### GET /plan/exec

Executa o próximo passo do plano ativo. Retorna `404` se não houver plano carregado, se todos os passos já tiverem sido consumidos, ou se o plano foi interrompido.

---

### GET /plan/config

Retorna o plano completo ativo, independente de quantos passos já foram consumidos. Retorna `404` se não houver plano carregado.

**Resposta:**
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

## Targets do Makefile

```
make deps      Baixa e organiza as dependências Go
make build     Compila o binário em bin/server
make run       Executa o servidor localmente
make vet       Executa go vet
make lint      Executa golangci-lint
make ci        Executa vet, lint e testes (espelha o pipeline de CI)
make test      Roda testes com race detector
make coverage  Gera relatório de cobertura
make help      Lista todos os targets com suas descrições
```

---

## Licença

Copyright (c) 2026 time-trial contributors.

Este projeto está licenciado sob a **Apache License 2.0**.

Você pode usar, modificar e distribuir este software de acordo com os termos da licença.

Consulte o arquivo [`LICENSE`](LICENSE) para os termos completos ou acesse
[apache.org/licenses/LICENSE-2.0](https://www.apache.org/licenses/LICENSE-2.0).
