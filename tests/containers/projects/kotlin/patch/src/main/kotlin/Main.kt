import java.io.File

fun main() {
    // Read version from build.gradle.kts
    val buildFile = File("build.gradle.kts").readText()
    val versionMatch = Regex("""version\s*=\s*"([^"]+)"""").find(buildFile)
    val version = versionMatch?.groupValues?.get(1) ?: "unknown"
    println("Version: $version")
}
