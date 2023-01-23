package translator

import "github.com/blastrain/vitess-sqlparser/sqlparser"

type SqlHandler interface {
	NeedHandleCreate(sql string, stat *sqlparser.CreateTable) (bool, error)
	HandleCreate(sql string, stat *sqlparser.CreateTable) (string, error)
}

type DefaultSqlHandler struct {
}

func (DefaultSqlHandler) NeedHandleCreate(sql string, stat *sqlparser.CreateTable) (bool, error) {
	return false, nil
}

func (DefaultSqlHandler) HandleCreate(sql string, stat *sqlparser.CreateTable) (string, error) {
	return sql, nil
}
