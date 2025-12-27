// Test application that reads version from JAR manifest
public class Main {
    public static void main(String[] args) {
        String version = Main.class.getPackage().getImplementationVersion();
        if (version == null) {
            version = "unknown";
        }
        System.out.println("Version: " + version);
    }
}
