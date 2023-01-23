package translator

import (
	"errors"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

type SqlType int

const (
	SqlTypeCreate SqlType = iota + 1
	SqlTypeDescribe
	SqlTypeSet
	SqlTypeShow
	//TODO need add support for DROP/Alter
	SqlTypeDelete SqlType = iota + 6 // from 10
	SqlTypeInsert
	SqlTypeUpdate
	SqlTypeSelect
)

var (
	ErrUnsupportedSQL = errors.New("unsupported sql")
)

type Req struct {
	SqlType
	RawSql        string
	TranslatedSql string

	statement sqlparser.Statement
	//TODO parameters
}

func ParseSQL(sql string, sh SqlHandler) (*Req, error) {
	//FIXME set command is not supported by current sqlparse library
	// So here manually exclude the support of this command
	var tlen = len(sql)
	if tlen > 10 {
		tlen = 10 // just to optimize the operation of toLowerCase
	}
	if strings.HasPrefix(strings.TrimSpace(strings.ToLower(sql[0:tlen])), "set") {
		return nil, ErrUnsupportedSQL
	}

	statement, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}

	switch stat := statement.(type) {
	case *sqlparser.CreateTable:
		required, err := sh.NeedHandleCreate(sql, stat)
		if err != nil {
			return nil, err
		}
		var translatedSql string
		if required {
			newSql, err := sh.HandleCreate(sql, stat)
			if err != nil {
				return nil, err
			}
			translatedSql = newSql
		} else {
			translatedSql = sql
		}
		return &Req{
			SqlType:       SqlTypeCreate,
			RawSql:        sql,
			TranslatedSql: translatedSql,
			statement:     statement,
		}, nil
	}

	return nil, ErrUnsupportedSQL
}
