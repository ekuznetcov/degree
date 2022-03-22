package rwmsr

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

//Read читает заданный msr-регистр на процессоре
//returns the uint64
func (d MSRDev) Read(msr int64) (uint64, error) {
	regBuf := make([]byte, 8)

	rc, err := syscall.Pread(d.fd, regBuf, msr)

	if err != nil {
		return 0, err
	}

	if rc != 8 {
		return 0, fmt.Errorf("Read wrong count of bytes: %d", rc)
	}

	//LittleEndian так как x86 процессор little endian
	msrOut := binary.LittleEndian.Uint64(regBuf)

	return msrOut, nil
}

//ReadMSRWithLocation как ReadMSR(), дополнительно принимает пользовательский путь,
//Для возможности использовать сторонний драйвер, например llnl/msr-safe
//На вход принимает строку содержащую %d для CPU. Пример /dev/cpu/%d/msr_safe
func ReadMSRWithLocation(cpu int, msrAddr int64, fmtStr string) (uint64, error) {
	msr, err := MSRWithLocation(cpu, fmtStr)
	if err != nil {
		return 0, err
	}

	msrD, err := msr.Read(msrAddr)
	if err != nil {
		return 0, err
	}

	return msrD, msr.Close()
}

//ReadMSR читает msr-регистр на заданном процессоре как разовую операцию
//cpu - процессор, msrAddr - адрес регистра
func ReadMSR(cpu int, msrAddr int64) (uint64, error) {
	msr, err := MSR(cpu)
	if err != nil {
		return 0, err
	}

	msrData, err := msr.Read(msrAddr)
	if err != nil {
		return 0, err
	}

	return msrData, msr.Close()
}
