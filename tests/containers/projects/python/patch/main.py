#!/usr/bin/env python3
"""Test application that reads version from pyproject.toml."""

import tomllib
from pathlib import Path

pyproject = Path("pyproject.toml")
with open(pyproject, "rb") as f:
    config = tomllib.load(f)

version = config["project"]["version"]
print(f"Version: {version}")
