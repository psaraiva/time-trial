# Arquitetura

[![Language: English](https://img.shields.io/badge/Idioma-English-blue?style=flat-square)](./ARCHITECTURE.md)

## Intenção

Este documento explica as escolhas arquiteturais deste projeto. A estrutura é intencionalmente plana e minimalista — não incompleta. Se parecer mais simples do que o esperado, esse é exatamente o objetivo.

O propósito do `time-trial` é estreito e bem definido: expor endpoints HTTP para injetar falhas controladas (status codes e delays) em serviços dependentes durante testes. Esse escopo não justifica camadas, abstrações ou padrões que acrescentariam complexidade sem agregar valor.

## Estrutura

```
cmd/
└── main.go          # Ponto de entrada: apenas injeção e wiring — sem lógica

internal/
├── entities/        # Estado de domínio: State e Plan
├── handlers/        # Camada HTTP: um arquivo por responsabilidade
└── middleware/      # Transversal: envelope de resposta
```

Três camadas e nada mais.

## Decisões deliberadas

**Sem camada de serviço.**
Os handlers consomem as entidades diretamente. Não há lógica de negócio complexa o suficiente para justificar uma camada de use-case ou service. Adicioná-la seria indireção pela indireção.

**Sem repositório ou camada de persistência.**
O estado é mantido em memória. A aplicação é stateless entre reinicializações por design — não há cenário nessa ferramenta onde estado durável agregue valor.

**Sem interfaces nas entidades.**
`State` e `Plan` são tipos concretos injetados nos handlers. Abstraí-los por trás de interfaces custaria legibilidade em favor de uma testabilidade que não se justifica aqui — as próprias entidades são simples o suficiente para serem testadas diretamente.

**Arquivos de handler planos.**
Cada arquivo de handler corresponde a um grupo de rotas. Não há classes base, hierarquias de handlers, nem lógica compartilhada entre handlers além do que o middleware já cobre.

**Dependência única.**
A única dependência externa é o [Fiber](https://gofiber.io/) para HTTP. Esta é uma restrição deliberada, não uma omissão.

## Modelo de concorrência

`State` usa `sync/atomic.Int32` em todos os campos — leituras e escritas lock-free sem contenção sob acesso concorrente.

`Plan` usa `sync.Mutex` para proteger o slice de states e o cursor, combinado com `atomic.Bool` para o flag de cancelamento. O mutex garante o acesso ordenado à sequência; o atomic trata o sinal de interrupção de forma independente.

## Quando esta arquitetura deve evoluir

Essa estrutura é adequada enquanto o escopo do projeto permanecer focado. Se o projeto crescer de formas que introduzam complexidade genuína — múltiplos domínios independentes, integrações externas, necessidade de persistência ou aumento significativo de regras de negócio — um layout orientado a domínio seria o próximo passo natural.

O sinal para reestruturar não é a quantidade de arquivos. É quando a estrutura atual começa a esconder relacionamentos ou tornar mudanças mais difíceis do que deveriam ser.

Até lá, a quantidade certa de estrutura é a mínima necessária para o código ser claro e o comportamento correto.
