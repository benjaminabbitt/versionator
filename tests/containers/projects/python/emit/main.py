#!/usr/bin/env python3
"""Test application that imports the generated version file."""

from _version import __version__, __version_tuple__

print(f"Version: {__version__}")
print(f"Version tuple: {__version_tuple__}")
