fn main() {
    // VERSION will be set by the compiler during build via environment variable
    let version = option_env!("VERSION").unwrap_or("0.0.0");
    
    println!("Sample Rust Application");
    println!("Version: {}", version);
}