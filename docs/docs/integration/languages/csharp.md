---
title: C#
description: Embed version in C# applications
sidebar_position: 7
---

# C\#

**Location:** [`examples/csharp/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/csharp)

C# generates a `Version.cs` static class at build time using `versionator output emit`:

```csharp title="examples/csharp/Program.cs"
using Version;

class Program
{
    static void Main(string[] args)
    {
        Console.WriteLine("Sample C# Application");
        Console.WriteLine($"Version: {VersionInfo.Version}");
    }
}
```

```makefile title="examples/csharp/Makefile (excerpt)"
version-file:
    versionator output emit csharp --output Version.cs

build: version-file
    dotnet build -c Release -o out
```

## Run it

```bash
$ cd examples/csharp && just run
Generating Version.cs using versionator emit...
Building C# application...
Build completed: out/SampleApp.dll
dotnet out/SampleApp.dll
Sample C# Application
Version: 0.0.16
```

## Source Code

- [`Program.cs`](https://github.com/benjaminabbitt/versionator/blob/master/examples/csharp/Program.cs)
- [`SampleApp.csproj`](https://github.com/benjaminabbitt/versionator/blob/master/examples/csharp/SampleApp.csproj)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/csharp/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/csharp/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/csharp/Containerfile)
