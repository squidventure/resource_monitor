package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	WindowsPrefix = "windows"
	LinuxPrefix   = "linux"
)

var (
	ResourceMonitorThreadPace = time.Second

	TheCPUMonitor    = &CPUMonitor{}
	TheMemoryMonitor = &MemoryMonitor{}
)

func (m MemoryMonitor) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// total, free, available
	start.Attr = m.Memory.XMLAttr()
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if err := e.EncodeToken(xml.EndElement{start.Name}); err != nil {
		return err
	}
	return e.Flush()

}

type MemoryMonitor struct {
	Memory
	mutex sync.RWMutex
}

func (c *MemoryMonitor) Update(memory Memory) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Memory = memory
}

type CPUMonitor struct {
	CPU
	mutex sync.RWMutex
}

func (c CPUMonitor) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = c.CPU.XMLAttr()
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if err := e.EncodeToken(xml.EndElement{start.Name}); err != nil {
		return err
	}
	return e.Flush()
}

func (c *CPUMonitor) Recalculate(cpu CPU) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	idleTicks := float64(cpu.Idle - c.Idle)
	totalTicks := float64(cpu.Total - c.Total)
	if totalTicks <= 0 {
		return
	}
	c.Idle = cpu.Idle
	c.Total = cpu.Total
	c.Utilization_percent = 100 * (totalTicks - idleTicks) / totalTicks
}

type Memory struct {
	Total, Free, Available uint64
}

func (m *Memory) XMLAttr() []xml.Attr {
	return []xml.Attr{
		xml.Attr{Name: xml.Name{"", "Total_kB"}, Value: fmt.Sprintf("%d", m.Total)},
		xml.Attr{Name: xml.Name{"", "Available_kB"}, Value: fmt.Sprintf("%d", m.Available)},
		xml.Attr{Name: xml.Name{"", "Free_kB"}, Value: fmt.Sprintf("%d", m.Free)},
	}
}

type CPU struct {
	Idle, Total         uint64
	Utilization_percent float64
}

func (c *CPU) XMLAttr() []xml.Attr {
	return []xml.Attr{
		xml.Attr{Name: xml.Name{"", "Total_ticks"}, Value: fmt.Sprintf("%d", c.Total)},
		xml.Attr{Name: xml.Name{"", "Idle_ticks"}, Value: fmt.Sprintf("%d", c.Idle)},
		xml.Attr{Name: xml.Name{"", "Utilization_percent"}, Value: fmt.Sprintf("%f", c.Utilization_percent)},
	}
}

func ContextNotDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func UpdateCPUStats() {
	goos := runtime.GOOS
	if strings.HasPrefix(goos, WindowsPrefix) {
		cpu := ReadWindowsCPUUsage()
		fmt.Println(cpu)
		TheCPUMonitor.Recalculate(cpu)
	} else if strings.HasPrefix(goos, LinuxPrefix) {
		cpu := ReadLinuxCPUUsage()
		TheCPUMonitor.Recalculate(cpu)
	} else {
		log.Fatalln("error: unsupported OS found", goos, "- cannot proceed")
	}
}

func UpdateMemoryStats() {
	goos := runtime.GOOS
	if strings.HasPrefix(goos, WindowsPrefix) {
		memory := ReadWindowsMemoryUsage()
		fmt.Println(memory)
		TheMemoryMonitor.Update(memory)
	} else if strings.HasPrefix(goos, LinuxPrefix) {
		memory := ReadLinuxMemoryUsage()
		TheMemoryMonitor.Update(memory)
	} else {
		log.Fatalln("error: unsupported OS found", goos, "- cannot proceed")
	}
}

func ResourceMonitorThread(ctx context.Context) {
	ticker := time.NewTicker(ResourceMonitorThreadPace)
	for ContextNotDone(ctx) {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			UpdateCPUStats()
			UpdateMemoryStats()
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go ResourceMonitorThread(ctx)
	done := make(chan os.Signal, 1)
	defer close(done)
	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
	cancel()
}
