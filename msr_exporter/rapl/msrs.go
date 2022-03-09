package rapl

// Читает RAPL регистры с процессоров (>sandybridge) Intel и AMD(Zen)

//   Существует три способа для чтения информации с доменов RAPL:
//  	1. Читать msr-регистры напрямую с /dev/cpu/??/msr
//  	2. Через perf_event_open() interface
//  	3. Читать с помощью sysfs powercap

//   Для чтения сырых данных из msr-регистров /dev/cpu/%d/msr драйвер должен быть установлен и
//  	должно быть установлено разрешение на чтение.
//  	Возможно надо будет выполнить команду "modprobe msr" перед началом работы.

//   perf_event_open() support requires at least Linux 3.14 and to have
//  	/proc/sys/kernel/perf_event_paranoid <
//	 не поддерживается AMD процессорами

//   sysfs powercap interface доступен из ядра с
//  	2d281d8196e38dd (3.13)
//	 не поддерживается AMD процессорами

//DomainMSR определяет адреса для различных msr-регистров
type DomainMSRs struct {
	PowerLimit   int64
	EnergyStatus int64
	Policy       int64
	PerfStatus   int64
	PowerInfo    int64
}

//Различные домены RAPL

//RAPLDomain - это стуктура определяющая различные домены RAPL
type RAPLDomain struct {
	Mask uint
	Name string
	MSRs DomainMSRs
}

//Package
var Package = RAPLDomain{0x1, "Package", DomainMSRs{0x610, 0x611, 0x0, 0x613, 0x614}}

//DRAM
var DRAM = RAPLDomain{0x2, "DRAM", DomainMSRs{0x618, 0x619, 0x0, 0x61b, 0x61c}}

//PP0
var PP0 = RAPLDomain{0x4, "PP0", DomainMSRs{0x638, 0x639, 0x63a, 0x63b, 0x0}}

//PP1
var PP1 = RAPLDomain{0x8, "PP1", DomainMSRs{0x640, 0x641, 0x642, 0x0, 0x0}}

//PSys

var PSys = RAPLDomain{0x16, "PSys", DomainMSRs{0x0, 0x64d, 0x0, 0x0, 0x0}}

//MSRPowerUnit определяет MSR для счетчика MSR_RAPL_POWER_UNIT
const MSRPowerUnit int64 = 0x606

// struct defs
//PowerLimitSetting устанавливает лимит мощности в установленное окно времени
type PowerLimitSetting struct {
	//Устанавливает среднее энергопотребление в Вт
	PowerLimit float64
	//Флаг активации ограничения энергопотребления
	EnableLimit bool
	//флаг активации ограничения частоты проццессора
	ClampingLimit bool
	//Временное окно в секундах
	TimeWindowLimit float64
}

//RAPLPowerLimit содержит данные в MSR_[DOMAIN]_POWER_LIMIT MSR
//Этот MSR регистр содержит два ограничения мощности из SDM:
//"Могут быть указаны два предельных значения мощности, соответствующие временным окнам различных размеров"
//"Каждый предел мощности обеспечивает независимый контроль ограничения, который позволит ядрам процессоров выйти ниже запрашиваемого уровня мощности."
type RAPLPowerLimit struct {
	Limit1 PowerLimitSetting
	Limit2 PowerLimitSetting
	Lock   bool
}

//RAPLPowerUnit осдержит данные в MSR_RAPL_POWER_UNIT MSR
type RAPLPowerUnit struct {
	//PowerUnits мультипликатор связанных с энергопотреблением данных в ватах
	PowerUnits float64
	//EnergyStatusUnits мультипликатор связанных с энергопотреблением данных в джоулях
	EnergyStatusUnits float64
	//TimeUnits мультипликатор связанных с временем данных в секундах
	TimeUnits float64
}

//RAPLPowerInfo содержит данные из MSR_[DOMAIN]_POWER_INFO MSR
type RAPLPowerInfo struct {
	ThermalSpecPower float64
	MinPower         float64
	MaxPower         float64
	MaxTimeWindow    float64
}
