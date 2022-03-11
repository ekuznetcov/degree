package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"os"
	"time"
)

type Duration struct {
	Start time.Time
	End   time.Time
}

type AggMetrics struct {
	Max float64
	Min float64
	Mid float64
}

type Node struct {
	nid  uint8
	node string
}

//getNodeAggMetrics возвращает агрегированные метрики(min, max, mean) для узлаа
//hostname string - сетевое имя узла, с к оторого собирались метрики
//duration Duration - структура содержащая начала и конец выполнения задания
func getNodeAggMetrics(hostname string, duration Duration) (AggMetrics, error) {
	var aggMetrics AggMetrics
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating clean: %v\n", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, "up", time.Now())
	if err != nil {
		fmt.Printf("Error quering Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
	return aggMetrics, nil
}

//writeToPostgres пишет агрегированные метрики для запущенного задания в Postgres
//nodeID int64 - id узла кластера в базе данных
//runID int64 - id запуска задания
//metrics AggMetrics - агрегированные значения метрик (max,min,mean)
//duration Duration - структура содержащая начала и конец выполнения задания
func writeToPostgres(nodeID int64, runID int64, metrics AggMetrics, duration Duration) {

}
func main() {
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %b\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r := v1.Range{
		Start: time.Now().Add(-time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}
	result, warnings, err := v1api.QueryRange(ctx)
}
func main01() {
	connStr := "root:root@tcp(172.19.0.3:5432)/cluster_db"
	conn, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	rows, err := conn.Query("select node, nid from mvs10p.nodes")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QuerryRow failed %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		n := Node{}
		err := rows.Scan(&n.node, &n.nid)
		if err != nil {
			fmt.Println(err)
			continue
		}
		nodes = append(nodes, n)
	}

	for _, n := range nodes {
		fmt.Println(n.node, n.nid)
	}
}
