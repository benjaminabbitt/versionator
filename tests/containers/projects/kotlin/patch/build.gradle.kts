plugins {
    kotlin("jvm") version "1.9.22"
    application
}

version = "0.0.0"

repositories {
    mavenCentral()
}

application {
    mainClass.set("MainKt")
}
