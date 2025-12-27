// Version injected via environment variable at build time
const VERSION: &str = env!("VERSION");

fn main() {
    println!("Version: {}", VERSION);
}
