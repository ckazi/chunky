package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	host       string
	port       int
	user       string
	password   string
	dbname     string
	table      string
	column     string
	limit      int
	offset     int
	maxGo      int
	file       string
	deleteFlag bool
)

func init() {
	flag.StringVar(&host, "h", "localhost", "Host address")
	flag.IntVar(&port, "p", 5432, "Port number")
	flag.StringVar(&user, "U", "postgres", "Database user")
	flag.StringVar(&password, "pwd", "", "Password for database")
	flag.StringVar(&dbname, "dbname", "", "Database name")
	flag.StringVar(&table, "table", "", "Database table")
	flag.StringVar(&column, "c", "id", "Column to order data")
	flag.IntVar(&limit, "limit", 5000, "Limit of rows per query")
	flag.IntVar(&offset, "offset", 0, "Offset to start from")
	flag.IntVar(&maxGo, "threads", 8, "Number of threads to use")
	flag.StringVar(&file, "file", "result.txt", "File to store results")
	flag.BoolVar(&deleteFlag, "del", false, "Delete flag to remove corrupted rows")
	flag.Parse()

	if dbname == "" || table == "" {
		fmt.Println("Usage of the program:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var tableLen int
	err = pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&tableLen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get table length: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sem := make(chan struct{}, maxGo)
	for i := offset; i < tableLen; i += limit {
		wg.Add(1)
		go func(l, o int) {
			sem <- struct{}{}
			defer wg.Done()
			defer func() { <-sem }()
			err := checkQuery(ctx, pool, l, o)
			if err != nil {
				fmt.Println("Error detected at offset:", o)
				findErrRow(ctx, pool, l, o)
			}
		}(limit, i)
	}
	wg.Wait()
}

func checkQuery(ctx context.Context, pool *pgxpool.Pool, l, o int) error {
	_, err := pool.Exec(ctx, fmt.Sprintf("SELECT * FROM %s ORDER BY %s LIMIT %d OFFSET %d", table, column, l, o))
	return err
}

func findErrRow(ctx context.Context, pool *pgxpool.Pool, lim, startOffset int) {
	if lim == 1 {
		uuid, err := getId(ctx, pool, column, table, startOffset)
		if err != nil {
			fmt.Println("Error getting UUID:", err)
			return
		}
		fmt.Println("Corrupted row id:", startOffset, "-", uuid)

		if deleteFlag {
			err := deleteEntryByUuid(ctx, pool, table, column, uuid)
			if err != nil {
				fmt.Println("Error deleting entry:", err)
				return
			}
			fmt.Println("Deleted entry id:", startOffset, "-", uuid)
			writeResultToFile(startOffset, uuid, true)
		} else {
			writeResultToFile(startOffset, uuid, false)
		}
		return
	}

	half := 0

	if lim%2 != 0 {
		half = lim/2 + 1
	} else {
		half = lim / 2
	}

	err := checkQuery(ctx, pool, half, startOffset)
	if err != nil {
		findErrRow(ctx, pool, half, startOffset)
	} else {
		findErrRow(ctx, pool, lim-half, startOffset+half)
	}
}

func getId(ctx context.Context, pool *pgxpool.Pool, column, table string, offset int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT 1 OFFSET %d", column, table, column, offset)).Scan(&id)
	return id, err
}

func deleteEntryByUuid(ctx context.Context, pool *pgxpool.Pool, table, column, uuid string) error {
	_, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s=$1", table, column), uuid)
	return err
}

func writeResultToFile(id int, uuid string, deleted bool) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	var resultString string
	if deleted {
		resultString = fmt.Sprintf("Corrupted row id: %d - %s - deleted\n", id, uuid)
	} else {
		resultString = fmt.Sprintf("Corrupted row id: %d - %s\n", id, uuid)
	}

	_, err = f.WriteString(resultString)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
