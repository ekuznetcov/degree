package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Stats struct {
	Max float64
	Min float64
	Avg float64
}

type Node struct {
	Name string
	Stat map[string]float64
}

type TaskMetrics struct {
	Mtid     int
	Nid      int
	Mid      int
	Rid      int
	MinValue float64 `db:"min_value"`
	MinTime  string  `db:"min_time"`
	MaxValue float64 `db:"max_value"`
	MaxTime  string  `db:"max_time"`
	AvgValue int     `db:"avg_value"`
}

//getNodesAggMetrics возвращает агрегированные метрики(min, max, mean) для узлаа
//nodeList []string - сетевое имя узл, с к отор собирались метрики
//dtimeStart time.Time
//timeEnd time.Time
func getNodesAggMetrics(nodeList []string, timeStart time.Time, timeEnd time.Time) ([]Node, error) {
	//формирование запроса
	stats := [3]string{"min", "max", "avg"}
	duration := timeEnd.Sub(timeStart).Round(time.Second).String()

	nodeStatList := make([]Node, 0)
	for _, node := range nodeList {
		nodeStat := Node{node, make(map[string]float64)}
		for _, stat := range stats {
			query := fmt.Sprintf("%s_over_time(irate(msr_rapl_joules_total{domain=~'Package|DRAM', instance='%s'}[2s])[%s:1s])", stat, node+":9876", duration)
			fmt.Println(query)
			//выполнение запроса
			client, err := api.NewClient(api.Config{
				Address: os.Getenv("PROM_URL"),
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
			total := 0
			for _, val := range result.(model.Vector) {
				total += int(val.Value)
			}
			nodeStat.Stat[stat] = float64(total)
		}
		nodeStatList = append(nodeStatList, nodeStat)
	}
	fmt.Println(nodeStatList)
	return nodeStatList, nil
}

// writeToPostgres пишет агрегированные метрики для запущенного задания в Postgres
// node Node - structer which contains node name and energy stat for run with runId
// systemName string - name logical part of system
// runId int64 - id запуска задания

func writeToPostgres(node Node, systemName string, runId int64) {
	db, err := sqlx.Connect("pgx", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	tx := db.MustBegin()
	query := fmt.Sprintf("INSERT INTO %s.task_metrics (rid, mid, min_value, avg_value, max_value) VALUES ($1, $2, $3, $4, $5)", systemName)
	tx.MustExec(query, runId, 5, node.Stat["min"], node.Stat["avg"], node.Stat["max"]) // mid=5 так как это значение mid для энергопотребления
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

//bundleToList convert nodes bundle to node list
//nodeBundle string - string with nodes bundle
func bundleToList(nodeBundle string) ([]string, error) {
	nodeList := []string{} //пример свертки узлов [20-25,30-32,40]
	nodeBundle = strings.Replace(nodeBundle, "[", "", -1)
	nodeBundle = strings.Replace(nodeBundle, "]", "", -1)
	nodeRanges := strings.Split(nodeBundle, ",")
	for _, nr := range nodeRanges {
		if strings.Contains(nr, "-") {
			rangeBound := strings.Split(nr, "-")
			start, err := strconv.ParseInt(rangeBound[0], 10, 64)
			if err != nil {
				return nil, err
			}
			end, err := strconv.ParseInt(rangeBound[1], 10, 64)
			if err != nil {
				return nil, err
			}
			for i := start; i < end; i++ {
				nodeList = append(nodeList, "node"+strconv.FormatInt(i, 10))
			}
		} else {
			nodeList = append(nodeList, "node"+nr)
		}
	}
	return nodeList, nil
}

func main() {
	nodeBundle := os.Args[1]
	nodeNameList, err := bundleToList(nodeBundle)
	if err != nil {
		log.Fatal(err)
	}
	timeStart, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	timeEnd, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	system := os.Args[4]
	taskId, err := strconv.ParseInt(os.Args[5], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	statsList, err := getNodesAggMetrics(nodeNameList, timeStart, timeEnd)
	if err != nil {
		log.Fatal(err)
	}
	for _, stat := range statsList {
		writeToPostgres(stat, system, taskId)
	}
}
