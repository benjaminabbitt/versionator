use std::fs;

fn main() {
    // Read version from Cargo.toml
    let cargo_toml = fs::read_to_string("Cargo.toml").expect("Failed to read Cargo.toml");
    let parsed: toml::Value = cargo_toml.parse().expect("Failed to parse Cargo.toml");
    let version = parsed["package"]["version"].as_str().unwrap_or("unknown");
    println!("Version: {}", version);
}
