"""Sample application entry point."""

from . import VERSION, GIT_HASH, BUILD_DATE


def main():
    print("Sample Python Application (Custom Template)")
    print(f"Version: {VERSION}")
    print(f"Git: {GIT_HASH}")
    print(f"Built: {BUILD_DATE}")


if __name__ == "__main__":
    main()
