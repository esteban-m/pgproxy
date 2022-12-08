package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/goodplayer/pgproxy/api"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	sqlFile string
	pgAddr  string
	dryRun  bool
)

func init() {
	flag.StringVar(&sqlFile, "file", "", "-file=sql.log : sql file that captured by sniffer")
	flag.StringVar(&pgAddr, "postgres", "", "-postgres=postgres://username:password@localhost:5432/database_name")
	flag.BoolVar(&dryRun, "dryrun", true, "-dryrun=false")

	flag.Parse()

	if len(sqlFile) == 0 {
		panic(errors.New("need specify sql file"))
	}
	if len(pgAddr) == 0 {
		panic(errors.New("need to specify postgres address"))
	}
}

func main() {

	fmt.Println(pgAddr)
	fmt.Println(sqlFile)

	cfg, err := pgxpool.ParseConfig(pgAddr)
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	f, err := os.Open(sqlFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// process file
	committed := false
	tx, err := pool.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() {
		if !committed {
			tx.Rollback(context.Background())
		}
	}()
	for {
		b := bufio.NewReader(f)
		line, err := b.ReadString('\n')
		if len(line) > 0 {
			v := new(api.SnifferElement)
			err := json.Unmarshal([]byte(line), v)
			if err != nil {
				fmt.Println("error format json:", line)
				panic(err)
			}
			// process line
			fmt.Println("==============processing sql=================")
			fmt.Println(v.SQL)
			t, err := tx.Exec(context.Background(), v.SQL)
			if err != nil {
				fmt.Println("process sql error.", err)
				panic(err)
			} else {
				fmt.Println("result:", t)
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	if dryRun {
		tx.Rollback(context.Background())
		committed = true
	} else {
		tx.Commit(context.Background())
		committed = true
	}

}
