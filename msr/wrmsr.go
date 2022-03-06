package msr

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

//Write writes a given val to the provided register
func (d MSRDev) Write(addr int64, val uint64) error {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, val)

	count, err := syscall.Pwrite(d.fd, buf, addr)
	if err != nil {
		return err
	}
	if count != 8 {
		return fmt.Errorf("Write count not a uint64: %d", count)
	}

	return nil
}

//WriteMSRWithLocation is like WriteMSR(), but takes a custom location,
//for use with testing or 3rd party utilities like llnl/msr-safe
//It takes a string that has a `%d` format specifier for the cpu. For example: /dev/cpu/%d/msr_safe
func WriteMSRWithLocation(cpu int, msrAddr int64, val uint64, fmtStr string) error {
	msr, err := MSRWithLocation(cpu, fmtStr)
	if err != nil {
		return err
	}

	err = msr.Write(msrAddr, val)
	if err != nil {
		return err
	}

	return msr.Close()
}

//WriteMSR() writes the val in MSR on the given CPU as a one-time operation
func WriteMSR(cpu int, msrAddr int64, val uint64) error {
	msr, err := MSR(cpu)
	if err != nil {
		return err
	}

	err = msr.Write(msrAddr, val)
	if err != nil {
		return err
	}

	return msr.Close()
}
