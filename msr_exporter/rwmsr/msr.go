package rwmsr

import (
	"fmt"
	"syscall"
)

const defaultFmtStr = "/dev/cpu/%d/msr"

//MSRDev предоставляет обработчик для операция чтения и записи
//для разового чтения/записи MSR, rwmsr предоставляет функцию {Read, Write} MSR*()
type MSRDev struct {
	fd int
}

//Close закрывает соединение с MSR-регистром
func (dev MSRDev) Close() error {
	return syscall.Close(dev.fd)
}

//MSR предоставляет интерфейс для повторного считывания из MSR регистров
func MSR(cpu int) (MSRDev, error) {
	cpuDir := fmt.Sprintf(defaultFmtStr, cpu)
	fd, err := syscall.Open(cpuDir, syscall.O_RDWR, 777)
	if err != nil {
		return MSRDev{}, err
	}
	return MSRDev{fd: fd}, nil
}

//MSRWithLocation как MSR(), дополнительно принимает пользовательский путь,
//Для возможности использовать сторонний драйвер, например llnl/msr-safe
//На вход принимает строку содержащую %d для CPU. Пример /dev/cpu/%d/msr_safe
func MSRWithLocation(cpu int, fmtString string) (MSRDev, error) {
	cpuDir := fmt.Sprintf(fmtString, cpu)
	fd, err := syscall.Open(cpuDir, syscall.O_RDWR, 777)
	if err != nil {
		return MSRDev{}, err
	}
	return MSRDev{fd}, err
}
