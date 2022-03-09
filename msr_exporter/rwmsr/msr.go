package rwmsr

import (
	"fmt"
	"syscall"
)

const defaultFmtStr = "/dev/cpu/%d/msr"

//MSRDev provides a handler for frequent read/write operations
//for one-off MSR read/writes, gomsr provides {Read, Write} MSR*() functions
type MSRDev struct {
	fd int
}

//Close closes the connection to the MSR register
func (d MSRDev) Close() error {
	return syscall.Close(d.fd)
}

//MSR provides the interface for reccuring acces to MSR interface
func MSR(cpu int) (MSRDev, error) {
	cpuDir := fmt.Sprintf(defaultFmtStr, cpu)
	fd, err := syscall.Open(cpuDir, syscall.O_RDWR, 777)
	if err != nil {
		return MSRDev{}, err
	}
	return MSRDev{fd: fd}, nil
}

//MSRWithLocation is tha same as MSR(), but takes a custom location,
//for use with testing or 3rd party utilities like llnl/msr-safe
//It takes a string that has a %d format for the CPU. For example /dev/cpu/%d/msr_safe
func MSRWithLocation(cpu int, fmtString string) (MSRDev, error) {
	cpuDir := fmt.Sprintf(fmtString, cpu)
	fd, err := syscall.Open(cpuDir, syscall.O_RDWR, 777)
	if err != nil {
		return MSRDev{}, err
	}
	return MSRDev{fd}, err
}
