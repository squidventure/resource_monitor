package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	LinuxMemoryFile = "/proc/meminfo"
	LinuxCPUFile    = "/proc/stat"
)

func TrimPrefixSuffixAndSpace(s, prefix, suffix string) string {
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, suffix)
	return strings.TrimSpace(s)
}

func ReadLinuxMemoryUsage() (memory Memory) {
	f, err := os.ReadFile(LinuxMemoryFile)
	if err != nil {
		log.Fatalln("error reading", LinuxMemoryFile, err)
	}
	for _, row := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(row, "MemTotal:") {
			row = TrimPrefixSuffixAndSpace(row, "MemTotal:", "kB")
			n, err := strconv.ParseUint(row, 10, 64)
			if err != nil {
				log.Fatalln("error parsing MemTotal", err)
			}
			memory.Total = n
		} else if strings.HasPrefix(row, "MemFree:") {
			row = TrimPrefixSuffixAndSpace(row, "MemFree:", "kB")
			n, err := strconv.ParseUint(row, 10, 64)
			if err != nil {
				log.Fatalln("error parsing MemFree", err)
			}
			memory.Free = n
		} else if strings.HasPrefix(row, "MemAvailable:") {
			row = TrimPrefixSuffixAndSpace(row, "MemAvailable:", "kB")
			n, err := strconv.ParseUint(row, 10, 64)
			if err != nil {
				log.Fatalln("error parsing MemAvailable", err)
			}
			memory.Available = n
		}
	}
	return
}

func ReadLinuxCPUUsage() (cpu CPU) {
	f, err := os.ReadFile(LinuxCPUFile)
	if err != nil {
		log.Fatalln("error reading", LinuxCPUFile, err)
	}
	for _, row := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(row, "cpu ") {
			fields := strings.Fields(row)
			for i := 1; i < len(fields); i++ {
				n, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Fatalln("error parsing CPU field", i, err)
				}
				cpu.Total += n
				if i == 4 {
					cpu.Idle = n
				}
			}
			return
		}
	}
	return
}
