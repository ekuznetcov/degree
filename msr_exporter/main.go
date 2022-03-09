package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"msr_exporter/rapl"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

type openYaml struct {
	Address   string
	Interface string
}

func topoPkgCPUMap() (map[int][]int, error) {

	sysdir := "/sys/devices/system/cpu/"
	cpuMap := make(map[int][]int)

	files, err := ioutil.ReadDir(sysdir)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("cpu[0-9]+")

	for _, file := range files {
		if file.IsDir() && re.MatchString(file.Name()) {

			fullPkg := filepath.Join(sysdir, file.Name(), "/topology/physical_package_id")
			dat, err := ioutil.ReadFile(fullPkg)
			if err != nil {
				return nil, errors.Wrapf(err, "error reading file %s", fullPkg)
			}
			phys, err := strconv.ParseInt(strings.TrimSpace(string(dat)), 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "error parsing value from %s", fullPkg)
			}
			var cpuCore int
			_, err = fmt.Sscanf(file.Name(), "cpu%d", &cpuCore)
			if err != nil {
				return nil, errors.Wrapf(err, "error fetching CPU core value from string %s", file.Name())
			}
			pkgList, ok := cpuMap[int(phys)]
			if !ok {
				cpuMap[int(phys)] = []int{cpuCore}
			} else {
				pkgList = append(pkgList, cpuCore)
				cpuMap[int(phys)] = pkgList
			}

		}
	}

	return cpuMap, nil
}

func GetEnergyCounts(path string) (map[string]float64, error) {
	counter := make(map[string]float64)

	top, err := topoPkgCPUMap()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching CPU topology")
	}

	for pkg, _ := range top {
		handler, err := rapl.CreateNewHandler(pkg, path)
		if err != nil {
			return nil, errors.Wrap(err, "error creating handler")
		}

		domains := handler.GetDomains()

		for _, domain := range domains {
			counter[domain.Name], err = handler.ReadEnergyStatus(domain)
			if err != nil {
				return nil, errors.Wrapf(err, "error reading energy status on domain %s", domain.Name)
			}
		}
	}
	return counter, nil
}

func main() {
	var config openYaml
	yamlFile, err := ioutil.ReadFile("/etc/msr_exporter/msr_exporter.yml")
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)

	metrics := func(w http.ResponseWriter, r *http.Request) {
		domains, err := GetEnergyCounts("")
		if err != nil {
			fmt.Println(err)
		}

		var responseText string //Переменная под текст ответа для Prometheus
		for domain := range domains {
			responseText += fmt.Sprintf("msr_rapl_%s_joules_total{} %f\n ", domain, domains[domain])
			fmt.Println(responseText)
		}
		fmt.Fprintf(w, responseText)
	}

	http.HandleFunc("/metrics", metrics)
	log.Fatal(http.ListenAndServe(config.Address, nil))
}
