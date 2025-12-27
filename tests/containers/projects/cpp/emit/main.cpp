// Test application that includes generated version header
#include <iostream>
#include "version.hpp"

int main() {
    std::cout << "Version: " << VERSION << std::endl;
    std::cout << "Major: " << VERSION_MAJOR << ", Minor: " << VERSION_MINOR << ", Patch: " << VERSION_PATCH << std::endl;
    return 0;
}
