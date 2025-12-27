// Test application that includes generated version header
#include <stdio.h>
#include "version.h"

int main() {
    printf("Version: %s\n", VERSION);
    printf("Major: %d, Minor: %d, Patch: %d\n", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH);
    return 0;
}
