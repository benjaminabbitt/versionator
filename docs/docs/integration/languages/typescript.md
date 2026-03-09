---
title: TypeScript
description: Embed version in TypeScript applications
sidebar_position: 11
---

# TypeScript

**Location:** [`examples/typescript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/typescript)

TypeScript generates a typed `version.ts` module using `versionator output emit`:

```typescript title="examples/typescript/src/index.ts"
import { VERSION } from './version.js';

function main(): void {
    console.log('Sample TypeScript Application');
    console.log(`Version: ${VERSION}`);
}

main();
```

```makefile title="examples/typescript/Makefile (excerpt)"
version-file:
    versionator output emit ts --output src/version.ts

build: version-file
    npx tsc

run: build
    node dist/index.js
```

## Run it

```bash
$ cd examples/typescript && just run
Generating version.ts using versionator emit...
Version 0.0.16 written to src/version.ts
Building TypeScript package...
Build completed!
node dist/index.js
Sample TypeScript Application
Version: 0.0.16
```

## Source Code

- [`src/index.ts`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/src/index.ts)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/Containerfile)
