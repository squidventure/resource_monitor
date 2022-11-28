package main

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	WindowsGlobalmemoryStatusEx = "GlobalMemoryStatusEx"
	//wmic cpu get loadpercentage
	//LoadPercentage
	//12
)

var (
	Kernel32DLL                 = syscall.NewLazyDLL("kernel32.dll")
	WindowsGetCPULoadPowershell = []string{"wmic", "cpu", "get", "loadpercentage"}
)

var sizeofMemoryStatusEx = uint32(unsafe.Sizeof(MemoryStatusEx{}))

type MemoryStatusEx struct {
	length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func ReadWindowsMemoryUsage() (memory Memory) {
	memoryStatusEX := MemoryStatusEx{length: sizeofMemoryStatusEx}
	proc := Kernel32DLL.NewProc(WindowsGlobalmemoryStatusEx)
	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&memoryStatusEX)))
	if ret != 1 {
		log.Fatalln("error calling", WindowsGlobalmemoryStatusEx, ret, err)
	}
	memory.Available = memoryStatusEX.AvailPhys
	memory.Free = memoryStatusEX.AvailPhys
	memory.Total = memoryStatusEX.TotalPhys
	return
}

var sizeOfSystemProcessorInformation = uint32(unsafe.Sizeof(SystemProcessorInformation{}))

type SystemProcessorInformation struct {
	IdleTime, KernelTime, UserTime, DpcTime, InterruptTime int64
	InterruptCount                                         uint32
}

func ReadWindowsCPUUsage() (cpu CPU) {
	cmd := exec.Command(WindowsGetCPULoadPowershell[0], WindowsGetCPULoadPowershell[1:]...)
	buffer := new(bytes.Buffer)
	cmd.Stderr = buffer
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		log.Fatalln("error running wmic", err)
	}
	s := strings.TrimSpace(strings.TrimPrefix(buffer.String(), "LoadPercentage"))
	cpu.Utilization_percent, err = strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalln("error parsing utilization percent from wmic", err)
	}
	return

	return
}
