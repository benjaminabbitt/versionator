---
title: JavaScript
description: Embed version in JavaScript applications
sidebar_position: 10
---

# JavaScript

**Location:** [`examples/javascript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/javascript)

JavaScript generates a `version.js` module using `versionator emit`:

```javascript title="examples/javascript/src/index.js"
import { VERSION } from './version.js';

function main() {
    console.log('Sample JavaScript Application');
    console.log(`Version: ${VERSION}`);
}

main();
```

```makefile title="examples/javascript/Makefile (excerpt)"
version-file:
    versionator emit js --output src/version.js

run: version-file
    node src/index.js
```

## Run it

```bash
$ cd examples/javascript && just run
Generating version.js using versionator emit...
Version 0.0.16 written to src/version.js
node src/index.js
Sample JavaScript Application
Version: 0.0.16
```

## Source Code

- [`src/index.js`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/src/index.js)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/Containerfile)
