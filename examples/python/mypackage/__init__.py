"""Sample Python package demonstrating versionator integration."""

try:
    from ._version import __version__
except ImportError:
    __version__ = "0.0.0"  # Fallback if _version.py not generated yet
