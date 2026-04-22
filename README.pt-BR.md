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
[![Language: English](https://img.shields.io/badge/Idioma-English-blue?style=flat-square)](./README.md)

[![CI](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml/badge.svg)](https://github.com/psaraiva/time-trial/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/psaraiva/time-trial)](https://goreportcard.com/report/github.com/psaraiva/time-trial)
[![codecov](https://codecov.io/gh/psaraiva/time-trial/branch/main/graph/badge.svg)](https://codecov.io/gh/psaraiva/time-trial)

Aplicação Go minimalista que expõe endpoints para simular comportamento de serviços dependentes — forçando status codes HTTP específicos, delays e geração dinâmica de response body em testes ou integrações.

## Visão Geral

A aplicação expõe rotas HTTP construídas com [Fiber](https://gofiber.io/). É possível definir um status code forçado com range de delay (`POST /time-trial`), carregar um plano ordenado de configurações (`POST /plan`), configurar um schema de geração de response body (`POST /param-resp`), e disparar execuções que respeitam essas configurações (`GET /sabotage`, `GET /plan/sabotage`). O estado é mantido em memória com atomics lock-free e mutex, sendo seguro para uso concorrente.

Quando uma configuração `param-resp` está ativa e uma execução retorna **200**, o body da resposta é gerado dinamicamente conforme o schema configurado, substituindo os metadados padrão de sabotagem.

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

## Swagger UI

Inicie o servidor e acesse:

```
http://localhost:7777/swag/index.html
```

A interface interativa documenta todas as rotas disponíveis, corpos de requisição e schemas de resposta.

Para regenerar a especificação Swagger após alterar as anotações:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swag
```

## Rotas

| Método | Caminho              | Descrição                                                         |
|--------|----------------------|-------------------------------------------------------------------|
| POST   | `/time-trial`        | Define ou reseta o status code forçado e range de delay           |
| GET    | `/time-trial/config` | Retorna a configuração atual de sabotagem                         |
| GET    | `/sabotage`          | Executa uma requisição com a configuração ativa de sabotagem      |
| POST   | `/plan`              | Define ou limpa um plano ordenado de configurações de sabotagem   |
| GET    | `/plan/sabotage`     | Executa o próximo passo do plano ativo                            |
| GET    | `/plan/config`       | Retorna o plano completo ativo                                    |
| POST   | `/param-resp`        | Define ou limpa a configuração de geração de response body        |
| GET    | `/param-resp/config` | Retorna a configuração atual do param-resp                        |

---

### POST /time-trial

Define a configuração ativa de sabotagem. Sem corpo, reseta para comportamento aleatório.

**Body:**
```json
{
  "code":     500,
  "delayMin": 100,
  "delayMax": 900
}
```

| Campo      | Valores aceitos                         | Descrição                        |
|------------|-----------------------------------------|----------------------------------|
| `code`     | `0`, `200`, `400`, `500`                | Status code forçado (`0` = reset)|
| `delayMin` | `1`–`60000` ms, ou `0` para desativar   | Delay mínimo em milissegundos    |
| `delayMax` | `>= delayMin`, máximo `60000` ms        | Delay máximo em milissegundos    |

**Sem corpo** → reseta sabotage (`code=0`, delays zerados).

---

### GET /sabotage

Executa uma requisição simulada usando a configuração ativa de sabotagem. Se nenhuma sabotagem estiver ativa (`code=0`), responde com um código aleatório (200, 400 ou 500).

Quando o código resultante for **200** e houver uma configuração `param-resp` ativa, o body é gerado dinamicamente — veja [POST /param-resp](#post-param-resp).

---

### GET /time-trial/config

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

### POST /plan

Carrega um plano ordenado de sabotagens. Cada item é uma configuração completa de sabotagem (State). Os passos são consumidos em ordem a cada chamada de `GET /plan`.

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

### GET /plan/sabotage

Executa o próximo passo do plano ativo. Retorna `404` se não houver plano carregado, se todos os passos já tiverem sido consumidos, ou se o plano foi interrompido.

Quando o código resultante for **200** e houver uma configuração `param-resp` ativa, o body é gerado dinamicamente — veja [POST /param-resp](#post-param-resp).

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

### POST /param-resp

Configura o schema usado para gerar o body da resposta quando um `200` é retornado por `/sabotage` ou `/plan`. Sem corpo, limpa a configuração ativa.

Tipos de propriedades suportados: `string`, `int`, `float`, `uuid`, `string-funny`.

**Body:**
```json
{
  "statusCode": 200,
  "item": {
    "isColection": true,
    "quantity": 5,
    "properties": [
      {
        "name": "id",
        "type": "uuid",
        "isRequired": true,
        "propertyUUID": {
          "version": 4
        }
      }, {
        "name": "name",
        "type": "string-funny",
        "isRequired": true
      }, {
        "name": "code",
        "type": "string",
        "isRequired": true,
        "maxLength": 10,
        "minLength": 3,
        "propertyString": {
          "chars": "abcdefghijklmnopqrstuvxzABCDEFGHIJKLMNOPQRSTUVXZ"
        }
      }, {
        "name": "value",
        "type": "float",
        "isRequired": true,
        "maxLength": 7777,
        "minLength": 0,
        "propertyFloat": {
          "floatPrecision": 2,
          "isAcceptNegativeValue": false
        }
      }, {
        "name": "version",
        "type": "int",
        "isRequired": true,
        "maxLength": 10,
        "minLength": 0,
        "propertyInt": {
          "isAcceptNegativeValue": false
        }
      }
    ]
  }
}
```

**Campos raiz:**

| Campo        | Descrição                                                                                                  |
|--------------|------------------------------------------------------------------------------------------------------------|
| `statusCode` | Deve ser `200` (único valor suportado)                                                                     |
| `item`       | Descreve o item ou coleção a ser gerado                                                                    |

**Campos de `item`:**

| Campo          | Descrição                                                                                                |
|----------------|----------------------------------------------------------------------------------------------------------|
| `isColection`  | `true` → retorna array JSON `[]`; `false` → retorna objeto único `{}`                                    |
| `quantity`     | Total de itens a gerar (ignorado quando `isColection` é `false`)                                         |
| `properties`   | 1..N definições de propriedade                                                                           |

O campo `name` aceita apenas letras, dígitos, `_` e `-`. Campos com `isRequired: false` são incluídos na resposta com valor `null`.

### Tabela de propriedades suportadas

| Tipo            | `isRequired` | `minLength` | `maxLength` | Config              |
|-----------------|:------------:|:-----------:|:-----------:|---------------------|
| `string`        | ✅           | ✅          | ✅          | `propertyString`    |
| `int`           | ✅           | ✅          | ✅          | `propertyInt`       |
| `float`         | ✅           | ✅          | ✅          | `propertyFloat`     |
| `uuid`          | ✅           | ❌          | ❌          | `propertyUUID`      |
| `string-funny`  | ✅           | ❌          | ❌          | —                   |

**Campos de configuração personalizada:**

| Config             | Campo                   | Descrição                                                                     |
|--------------------|-------------------------|-------------------------------------------------------------------------------|
| `propertyString`   | `chars`                 | Conjunto de caracteres para geração (somente letras A–Z, a–z)                 |
| `propertyInt`      | `isAcceptNegativeValue` | Permite valores negativos na geração                                          |
| `propertyFloat`    | `floatPrecision`        | Total de casas decimais (0–5)                                                 |
| `propertyFloat`    | `isAcceptNegativeValue` | Permite valores negativos na geração                                          |
| `propertyUUID`     | `version`               | Versão do UUID: `1` (baseado em tempo), `4` (aleatório), `7` (ordenado)       |
| `string-funny`     | —                       | Sem config. Gera `<adjetivo>_<objeto>` (ex: `angry_spoon`, `sleepy_bucket`)   |

---

### GET /param-resp/config

Retorna a configuração atual do param-resp exatamente como foi submetida. Retorna `404` se nenhuma configuração estiver ativa.

---

## Targets do Makefile

```
make deps      Baixa e organiza as dependências Go
make build     Compila o binário em bin/server
make run       Executa o servidor localmente
make swag      Gera a documentação Swagger (requer o CLI swag)
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
