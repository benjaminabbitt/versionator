import version.Version;

public class Main {
    public static void main(String[] args) {
        System.out.println("Version: " + Version.VERSION);
        System.out.println("Major: " + Version.MAJOR + ", Minor: " + Version.MINOR + ", Patch: " + Version.PATCH);
    }
}
