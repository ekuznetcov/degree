package rapl

import (
	"fmt"

	"msr_exporter/rwmsr"

	"github.com/pkg/errors"
)

//RAPLHandler manages a stateful connection to the RAPL system.
type RAPLHandler struct {
	availDomains []RAPLDomain //Available RAPL domains
	domainMask   uint         //a bitmask to make it easier to find available domains
	msrDev       rwmsr.MSRDev
	units        RAPLPowerUnit
}

// ErrMSRDoesNotExist ощибка, возникающая если Dоmain не существует в RAPL
var ErrMSRDoesNotExist = errors.New("MSR does not exist on selected Domain")

//Берет эту информацию из ядра, проверяя Energy Staus MSR-регистров
func getAvailableDomains(cpu int, msr rwmsr.MSRDev) ([]RAPLDomain, uint) {
	var actualDomains []RAPLDomain
	var dm uint

	if _, exists := msr.Read(Package.MSRs.EnergyStatus); exists == nil {
		actualDomains = append(actualDomains, Package)
		dm = dm | Package.Mask
	}

	if _, exists := msr.Read(DRAM.MSRs.EnergyStatus); exists == nil {
		actualDomains = append(actualDomains, DRAM)
		dm = dm | DRAM.Mask
	}

	if _, exists := msr.Read(PP0.MSRs.Policy); exists == nil {
		actualDomains = append(actualDomains, PP0)
		dm = dm | PP0.Mask
	}

	if _, exists := msr.Read(PP1.MSRs.EnergyStatus); exists == nil {
		actualDomains = append(actualDomains, PP1)
		dm = dm | PP1.Mask
	}

	if _, exists := msr.Read(PSys.MSRs.EnergyStatus); exists == nil {
		actualDomains = append(actualDomains, PSys)
		dm = dm | PSys.Mask
	}

	return actualDomains, dm
}

//CreateNewHandler создает обработчик для регистра RAPL для заданного CPU
func CreateNewHandler(cpu int, fmtS string) (RAPLHandler, error) {

	var msr rwmsr.MSRDev
	var err error
	if fmtS == "" {
		msr, err = rwmsr.MSR(cpu)
		if err != nil {
			return RAPLHandler{}, errors.Wrap(err, "error creating MSR handler")
		}
	} else {
		msr, err = rwmsr.MSRWithLocation(cpu, fmtS)
		if err != nil {
			return RAPLHandler{}, errors.Wrapf(err, "error creating MSR handler with location %s", fmtS)
		}
	}

	domains, mask := getAvailableDomains(cpu, msr)
	if len(domains) == 0 {
		return RAPLHandler{}, fmt.Errorf("No RAPL domains available on CPU")
	}

	handler := RAPLHandler{availDomains: domains, domainMask: mask, msrDev: msr}

	handler.units, err = handler.ReadPowerUnit()
	if err != nil {
		return RAPLHandler{}, errors.Wrapf(err, "error reading power units")
	}

	return handler, nil
}

//ReadPowerLimit возвращает MSR_[DOMAIN]_POWER_LIMIT MSR
//Этот MSR регистр определяет ограничение мощности для заданного домена. Этот регистр есть в каждом домене
func (handler RAPLHandler) ReadPowerLimit(domain RAPLDomain) (RAPLPowerLimit, error) {
	if domain.Mask&handler.domainMask == 0 {
		return RAPLPowerLimit{}, fmt.Errorf("Domain %s does not exist on system", domain.Name)
	}

	data, err := handler.msrDev.Read(domain.MSRs.PowerLimit)
	if err != nil {
		return RAPLPowerLimit{}, err
	}

	var singleLimit = false
	if domain != Package {
		singleLimit = true
	}

	return parsePowerLimit(data, handler.units, singleLimit), nil
}

//ReadEnergyStatus возвращает MSR_[DOMAIN]_ENERGY_STATUS MSR
//Этот MSR представляет одно 32 битовое поле, предоставляюшее отчет об энергопотреблении домена
//Значение регистра обновляется ~1мс, каждый домен имеет этот регистр.
func (handler RAPLHandler) ReadEnergyStatus(domain RAPLDomain) (float64, error) {
	if (domain.Mask & handler.domainMask) == 0 {
		return 0, fmt.Errorf("Domain %s doesn't exists on system", domain.Name)
	}

	data, err := handler.msrDev.Read(domain.MSRs.EnergyStatus)
	if err != nil {
		return 0, err
	}

	return float64(data&0xffffffff) * handler.units.EnergyStatusUnits, nil
}

//ReadPolicy возвращает MSR_[DOMAIN]_POLICY MSR
//Этот MSR представляет 4 битовое поле, содержащее информацию о приоритете
//распределения энергопотребления между поддоменами в родительском домене
func (handler RAPLHandler) ReadPolicy(domain RAPLDomain) (uint64, error) {
	if (domain.Mask & handler.domainMask) == 0 {
		return 0, fmt.Errorf("Domain %s doesn't exists on system", domain.Name)
	}

	if domain.MSRs.Policy == 0 {
		return 0, ErrMSRDoesNotExist
	}

	data, err := handler.msrDev.Read(domain.MSRs.Policy)
	if err != nil {
		return 0, err
	}
	return data & 0x1f, nil
}

//ReadPerfStatus возвращает MSR_[DOMAIN]_PERF_STATUS MSR.
//Это значение предоставляет количество времени, которое домен был ограничен из-за ограничений RAPL
//Недоступно в PP1
func (handler RAPLHandler) ReadPerfStatus(domain RAPLDomain) (float64, error) {
	if domain.Mask&handler.domainMask == 0 {
		return 0, fmt.Errorf("Domain %s does not exist on system", domain.Name)
	}

	if domain.MSRs.PerfStatus == 0 {
		return 0, ErrMSRDoesNotExist
	}

	data, err := handler.msrDev.Read(domain.MSRs.PerfStatus)
	if err != nil {
		return 0, err
	}

	return float64(data&0xffffffff) * handler.units.TimeUnits, nil
}

//ReadPowerInfo возвращает MSR_[DOMAIN]_POWER_INFO MSR.
//Этот MSR недоступен в PP0/PP1 доменах
func (handler RAPLHandler) ReadPowerInfo(domain RAPLDomain) (RAPLPowerInfo, error) {
	if (domain.Mask & handler.domainMask) == 0 {
		return RAPLPowerInfo{}, fmt.Errorf("Domain %s does not exist on system", domain.Name)
	}

	if domain.MSRs.PerfStatus == 0 {
		return RAPLPowerInfo{}, ErrMSRDoesNotExist
	}

	data, err := handler.msrDev.Read(domain.MSRs.PowerInfo)
	if err != nil {
		return RAPLPowerInfo{}, err
	}

	return parsePowerInfo(data, handler.units), nil
}

//ReadPowerUnit вовращает MSR_RAPL_POWER_UNIT MSR
//Нет связаных доменов
func (handler RAPLHandler) ReadPowerUnit() (RAPLPowerUnit, error) {
	data, err := handler.msrDev.Read(MSRPowerUnit)
	if err != nil {
		return RAPLPowerUnit{}, nil
	}

	return parsePowerUnit(data), nil
}

// GetDomains возвращает список доступных доменов
func (handler RAPLHandler) GetDomains() []RAPLDomain {
	return handler.availDomains
}
