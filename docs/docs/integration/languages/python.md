---
title: Python
description: Embed version in Python packages
sidebar_position: 9
---

# Python

**Location:** [`examples/python/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/python)

Python uses `versionator emit` to generate a `_version.py` module:

```python title="examples/python/mypackage/main.py"
"""Sample application entry point."""

from . import __version__


def main():
    print("Sample Python Application")
    print(f"Version: {__version__}")


if __name__ == "__main__":
    main()
```

```makefile title="examples/python/Makefile (excerpt)"
version-file:
    versionator emit python --output mypackage/_version.py

run: version-file
    python -m mypackage.main
```

## Run it

```bash
$ cd examples/python && just run
Generating _version.py using versionator emit...
Version 0.0.16 written to mypackage/_version.py
python3 -m mypackage.main
Sample Python Application
Version: 0.0.16
```

## Source Code

- [`mypackage/main.py`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/mypackage/main.py)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/Containerfile)
