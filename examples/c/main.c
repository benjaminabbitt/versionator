#include <stdio.h>

// VERSION will be set by the compiler during build
#ifndef VERSION
#define VERSION "0.0.0"
#endif

int main() {
    printf("Sample C Application\n");
    printf("Version: %s\n", VERSION);
    return 0;
}