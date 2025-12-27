// Test application that uses preprocessor defines from build flags
#include <iostream>

#ifndef VERSION
#define VERSION "unknown"
#endif

int main() {
    std::cout << "Version: " << VERSION << std::endl;
    return 0;
}
