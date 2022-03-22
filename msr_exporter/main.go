package main

import (
	"escheduler/msr_exporter/rapl"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var msrCollector = NewMSRCollector()

type openYaml struct {
	Address   string
	Interface string
}

/*
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
*/

func getRAPLHandler(path string) (rapl.RAPLHandler, error) {
	handler, err := rapl.CreateNewHandler(0, path) //0 потому что я не нашел зависимости от того с какого ядра снимать информацию
	if err != nil {
		return handler, errors.Wrap(err, "error creating handler")
	}
	return handler, nil
}

func getDomainValue(domain rapl.RAPLDomain, handler rapl.RAPLHandler) (float64, error) {
	metricValue, err := handler.ReadEnergyStatus(domain)
	if err != nil {
		return -1, errors.Wrapf(err, "error reading energy status on domain %s", domain.Name)
	}
	return metricValue, nil
}

type MSRCollector struct {
	counterDesc *prometheus.Desc
}

func NewMSRCollector() *MSRCollector {
	return &MSRCollector{
		counterDesc: prometheus.NewDesc("msr_rapl_joules_total", "total energy consumption on this domain (J). If you want get power you need get joules per second",
			[]string{"domain"}, nil),
	}
}

func (collector *MSRCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.counterDesc
}

func (collector *MSRCollector) Collect(channel chan<- prometheus.Metric) {
	handler, err := getRAPLHandler("")
	if err != nil {
		log.Fatal(err)
	}

	domains := handler.GetDomains()
	for _, domain := range domains {
		value, err := getDomainValue(domain, handler)
		if err != nil {
			log.Fatal(err)
		}

		channel <- prometheus.MustNewConstMetric(
			collector.counterDesc,
			prometheus.CounterValue,
			value,
			domain.Name,
		)
	}
	handler.Close()

}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
		ch := make(chan prometheus.Metric)
		msrCollector.Collect(ch)
		close(ch)
	})
}

func main() {
	//hook in the collector
	prometheus.MustRegister(msrCollector)
	//create router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	//router.Use(prometheusMiddleware)
	//home page
	router.HandleFunc("/", func(w http.ResponseWriter, rr *http.Request) {
		w.Write([]byte(`<html>
		<body>
			<h1>MSR Exporter</h1>
		
			<p>msr_exporter - это программное средство сбора данных об энергопотреблении ЦПУ на ВС с ОС Linux.</p>
			<p><strong>поддерживаемые платформы</strong></p>
			<p>msr_exporter поддерживает все архитектуры процессоров поддерживающие интерфейс RAPL</p>
			<ul>
				<li>Intel начиная с архиетктуры Sandy Bridge</li>
				<li>AMD начиная с архитктуры Pyzen</li>
			</ul>
		</body>
		</html>`))
	})
	//metrics page
	router.Handle("/metrics", promhttp.Handler())
	ch := make(chan prometheus.Metric)
	go msrCollector.Collect(ch)
	//run http server
	fmt.Println("Serving requests on port 9100")
	err := http.ListenAndServe(":9100", router)
	log.Fatal(err)

}
