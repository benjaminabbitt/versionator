#include <iostream>

// VERSION will be set by the compiler during build
#ifndef VERSION
#define VERSION "0.0.0"
#endif

int main() {
    std::cout << "Sample C++ Application" << std::endl;
    std::cout << "Version: " << VERSION << std::endl;
    return 0;
}