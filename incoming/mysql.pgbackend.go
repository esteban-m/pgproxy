package incoming

import (
	"context"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgBackend struct {
	pool *pgxpool.Pool
}

func (b *PgBackend) HandleSelect(selectQuery string) (*mysql.Result, error) {
	// sql conversion
	outSql, mod, err := selectSQLModifications.Run(selectQuery)
	if err != nil {
		return nil, err
	}

	rows, err := b.pool.Query(context.Background(), outSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fieldDescriptions := rows.FieldDescriptions()
	var fields []string
	for _, v := range fieldDescriptions {
		fields = append(fields, string(v.Name))
	}
	fields = fields[:len(fields)-mod.AddedLastMetaColumnCnt] // exclude last meta columns which are not expected by client

	var results [][]interface{}
	for rows.Next() {
		r, err := rows.Values()
		if err != nil {
			return nil, err
		}
		results = append(results, r[:len(r)-mod.AddedLastMetaColumnCnt]) // exclude last meta columns which are not expected by client
	}

	// data mapping
	if r, err := selectResultSetDataMappings.Run(results); err != nil {
		return nil, err
	} else {
		results = r
	}

	resultSet, err := BuildSimpleResultset(fields, results, false) // have to use text format ?
	if err != nil {
		return nil, err
	}
	return &mysql.Result{
		Status:       0,
		Warnings:     0,
		InsertId:     0,
		AffectedRows: 0,
		Resultset:    resultSet,
	}, nil
}

func NewPgBackend(connString string) (*PgBackend, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &PgBackend{
		pool: pool,
	}, nil
}
