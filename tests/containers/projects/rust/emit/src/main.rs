mod version;

fn main() {
    println!("Version: {}", version::VERSION);
    println!("Major: {}, Minor: {}, Patch: {}", version::MAJOR, version::MINOR, version::PATCH);
}
