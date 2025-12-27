#!/usr/bin/env python3
"""Test application that reads version from setup.py."""

import re
from pathlib import Path

setup_py = Path("setup.py").read_text()
match = re.search(r'version\s*=\s*["\']([^"\']+)["\']', setup_py)
if match:
    version = match.group(1)
    print(f"Version: {version}")
else:
    print("Version not found in setup.py")
    exit(1)
