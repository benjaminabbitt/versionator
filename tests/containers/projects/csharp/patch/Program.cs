using System;
using System.Reflection;

class Program
{
    static void Main()
    {
        var version = Assembly.GetExecutingAssembly().GetName().Version;
        // .NET Version: Major.Minor.Build.Revision (Build = Patch in semver)
        Console.WriteLine($"Version: {version?.Major}.{version?.Minor}.{version?.Build}.{version?.Revision}");
        Console.WriteLine($"Major: {version?.Major}, Minor: {version?.Minor}, Patch: {version?.Build}, Revision: {version?.Revision}");
    }
}
