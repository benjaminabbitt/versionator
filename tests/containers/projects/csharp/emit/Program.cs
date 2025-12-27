// Test application that imports generated version class
using Version;

Console.WriteLine($"Version: {VersionInfo.Version}");
Console.WriteLine($"AssemblyVersion: {VersionInfo.AssemblyVersion}");
Console.WriteLine($"Major: {VersionInfo.Major}, Minor: {VersionInfo.Minor}, Patch: {VersionInfo.Patch}, Revision: {VersionInfo.Revision}");
Console.WriteLine($"GitHash: {VersionInfo.GitHash}");
