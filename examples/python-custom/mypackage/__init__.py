"""Sample Python package demonstrating versionator custom template integration."""

try:
    from ._version import VERSION, FULL_VERSION, GIT_HASH, BUILD_DATE
except ImportError:
    VERSION = "0.0.0"
    FULL_VERSION = "0.0.0"
    GIT_HASH = "unknown"
    BUILD_DATE = "unknown"
