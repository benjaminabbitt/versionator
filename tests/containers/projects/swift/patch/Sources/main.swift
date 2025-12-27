import Foundation

// Read version from Package.swift comment
func readVersionFromPackage() -> String {
    guard let content = try? String(contentsOfFile: "Package.swift", encoding: .utf8) else {
        return "unknown"
    }
    for line in content.components(separatedBy: .newlines) {
        if line.contains("// VERSION:") {
            let parts = line.components(separatedBy: ":")
            if parts.count >= 2 {
                return parts[1].trimmingCharacters(in: .whitespaces)
            }
        }
    }
    return "unknown"
}

let version = readVersionFromPackage()
print("Version: \(version)")
