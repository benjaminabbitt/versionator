// Test application that reads version from assembly metadata
// Version is injected via: dotnet build /p:Version=X.Y.Z
using System.Reflection;

var assembly = Assembly.GetExecutingAssembly();

// InformationalVersion contains full semver with prerelease/metadata
var infoVersion = assembly
    .GetCustomAttribute<AssemblyInformationalVersionAttribute>()
    ?.InformationalVersion ?? "unknown";

// Assembly.Version is the 4-component .NET version (Major.Minor.Build.Revision)
var asmVersion = assembly.GetName().Version;

Console.WriteLine($"InformationalVersion: {infoVersion}");
Console.WriteLine($"AssemblyVersion: {asmVersion?.Major}.{asmVersion?.Minor}.{asmVersion?.Build}.{asmVersion?.Revision}");
Console.WriteLine($"Revision: {asmVersion?.Revision}");
