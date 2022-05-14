package main

import (
	"context"
	"errors"
	"flag"
	"os"

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

	//TODO
}
