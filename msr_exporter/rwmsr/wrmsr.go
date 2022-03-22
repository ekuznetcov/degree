package rwmsr

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

//Write пишет заданное значение в заданный msr-регистр
//val - значение, addr - адрес регистра
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

//WriteMSRWithLocation как WriteMSR(), дополнительно принимает пользовательский путь,
//Для возможности использовать сторонний драйвер, например llnl/msr-safe
//На вход принимает строку содержащую %d для CPU. Пример /dev/cpu/%d/msr_safe
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

//WriteMSR() пишет значение в MSR-регистр на заданном процессоре как разовую операцию
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
