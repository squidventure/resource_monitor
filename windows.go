package main

const (
	Kernel32                               = "kernel32.dll"
	WindowsGetNumaAvailableMemory          = "GetNumaAvailableMemory"
	WindowsPhysicallyInstalledSystemMemory = "GetPhysicallyInstalledSystemMemory"
	WindowsGetSystemCpuSetInformation      = "GetSystemCpuSetInformation	"
)

func ReadWindowsMemoryUsage() Memory {

}
