// Test application that uses preprocessor defines from build flags
#include <stdio.h>

#ifndef VERSION
#define VERSION "unknown"
#endif

int main() {
    printf("Version: %s\n", VERSION);
    return 0;
}
