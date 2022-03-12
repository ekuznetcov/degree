package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"os"
	"time"
)

//getNodeAggMetrics возвращает агрегированные метрики(min, max, mean) для узлаа
//hostname string - сетевое имя узла, с к оторого собирались метрики
//duration Duration - структура содержащая начала и конец выполнения задания
func getNodeAggMetrics(hostname string, queryStr string, duration v1.Range) (model.Value, error) {
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

	result, warnings, err := v1api.QueryRange(ctx, queryStr, duration)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)

	return result, nil
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
	domains := [5]string{"psys", "package", "dram", "pp0", "pp1"}
	stats := [3]string{"min", "max", "avg"}
	hostname := "http://localhost:9090"
	for _, domain := range domains {
		for _, stat := range stats {
			query := fmt.Sprintf("%s(rate(node_cpu_energy_%s[15s]))", stat, domain)
			fmt.Println(query)
			duration := v1.Range{
				Start: time.Now().Add(-time.Minute * 30),
				End:   time.Now(),
				Step:  time.Minute,
			}
			getNodeAggMetrics(hostname, query, duration)
		}
	}
}
