package main

import (
	"C"
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"os"
	"time"
)

type Stats struct {
	Max model.Value
	Min model.Value
	Avg model.Value
}

type Domain struct {
	Name  string
	Stats Stats
}

//getNodeAggMetrics возвращает агрегированные метрики(min, max, mean) для узлаа
//hostname string - сетевое имя узла, с к оторого собирались метрики
//duration Duration - структура содержащая начала и конец выполнения задания
//export getNodeAggMetrics
func getNodeAggMetrics(serverAddr string, hostList []string, timeStart time.Time, timeEnd time.Time) (model.Value, error) {
	//формирование запроса
	domains := [5]string{"psys", "package", "dram", "uncore", "core"}
	stats := [3]string{"min", "max", "avg"}
	hostsReg := ""
	for _, host := range hostList {
		hostsReg += host + "|"
	}
	duration := timeEnd.Sub(timeStart).Round(time.Second).String()

	var query string
	for _, domain := range domains {
		for _, stat := range stats {
			query = fmt.Sprintf("max without (instance) (%s_over_time(irate(node_rapl_%s_joules_total{instance=~'%s'}[2s])[%s:1s]))", stat, domain, hostsReg, duration)
			fmt.Println(query)

			//выполнение запроса
			client, err := api.NewClient(api.Config{
				Address: serverAddr,
			})
			if err != nil {
				fmt.Printf("Error creating client: %b\n", err)
				os.Exit(1)
			}
			v1api := v1.NewAPI(client)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, warnings, err := v1api.Query(ctx, query, timeStart)
			if err != nil {
				fmt.Printf("Error querying Prometheus: %v\n", err)
				os.Exit(1)
			}
			if len(warnings) > 0 {
				fmt.Printf("Warnings: %v\n", warnings)
			}
			fmt.Printf("Result:\n%v\n", result)
		}
	}
	return nil, nil
}

//writeToPostgres пишет агрегированные метрики для запущенного задания в Postgres
//nodeID int64 - id узла кластера в базе данных
//runID int64 - id запуска задания
//metrics AggMetrics - агрегированные значения метрик (max,min,mean)
//duration Duration - структура содержащая начала и конец выполнения задания
//func writeToPostgres(nodeID int64, runID int64, metrics AggMetrics, duration Duration) {
//
//}

func main() {
	serverAddr := "http://localhost:9090"
	hostsList := []string{"localhost:9100"}

	getNodeAggMetrics(serverAddr, hostsList, time.Now().Add(-time.Hour), time.Now())
}

/*stats := [3]string{"min", "max", "avg"}
hostname := "http://localhost:9090"
for _, stat := range stats {
	query := fmt.Sprintf("%s_over_time(irate(node_cpu_seconds_total{cpu='0', mode=\"iowait\"}\n[30s])[1h:15s])", stat)
	fmt.Println(query)
	client, err := api.NewClient(api.Config{
		Address: hostname,
	})
	if err != nil {
		fmt.Printf("Error creating client: %b\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, query, time.Now().Add(-time.Hour))
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}*/
