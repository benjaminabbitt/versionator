---
title: Languages
description: Language-specific version embedding examples
sidebar_position: 0
---

# Language Integration

Embed version information into your applications. Choose your language:

## Compiled Languages

Inject version at compile time:

| Language | Mechanism |
|----------|-----------|
| [Go](./go) | `-ldflags` linker injection |
| [Rust](./rust) | `option_env!()` compile-time |
| [C](./c) | `-D` preprocessor defines |
| [C++](./cpp) | `-D` preprocessor defines |

## JVM Languages

Generate source files at build time:

| Language | Mechanism |
|----------|-----------|
| [Java](./java) | Template-generated source |
| [Kotlin](./kotlin) | `versionator emit kotlin` |

## .NET Languages

| Language | Mechanism |
|----------|-----------|
| [C#](./csharp) | `versionator emit csharp` |

## Apple Platforms

| Language | Mechanism |
|----------|-----------|
| [Swift](./swift) | `versionator emit swift` |

## Interpreted Languages

Generate version modules:

| Language | Mechanism |
|----------|-----------|
| [Python](./python) | `versionator emit python` |
| [JavaScript](./javascript) | `versionator emit js` |
| [TypeScript](./typescript) | `versionator emit ts` |
| [Ruby](./ruby) | `versionator emit ruby` |

## Containers

| Platform | Mechanism |
|----------|-----------|
| [Docker](./docker) | Build args + OCI labels |
