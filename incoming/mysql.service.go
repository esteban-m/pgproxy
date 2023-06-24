package incoming

import (
	"fmt"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"log"
	"net"
	"reflect"
	"strings"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/server"
)

type commandHandler struct {
	AllowedDatabase map[string]struct{}

	pgbackend *PgBackend
}

func (c *commandHandler) UseDB(dbName string) error {
	_, ok := c.AllowedDatabase[strings.ToLower(dbName)]
	if ok {
		return nil
	}

	log.Println(fmt.Errorf("use is not supported:[%s]", dbName))
	return fmt.Errorf("use is not supported:[%s]", dbName)
}

func (c *commandHandler) HandleQuery(query string) (*mysql.Result, error) {
	statement, err := sqlparser.Parse(query)
	if err != nil {
		return nil, err
	}
	log.Println(reflect.TypeOf(statement))
	switch v := statement.(type) {
	case *sqlparser.Select:
		collector := struct {
			Type                string
			SQL                 string
			IsAccessingVariable bool // select field with @ or @@
		}{
			Type: "select",
			SQL:  query,
		}
		// Samples:
		// #1 Not supported: SELECT @@SESSION.sql_mode
		{
			for _, vv := range v.SelectExprs {
				if err := vv.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
					if col, ok := node.(*sqlparser.ColName); ok {
						if strings.HasPrefix(col.Name.String(), "@") || strings.HasPrefix(col.Name.String(), "@@") ||
							strings.HasPrefix(col.Qualifier.Name.String(), "@") || strings.HasPrefix(col.Qualifier.Name.String(), "@@") {
							collector.IsAccessingVariable = true
						}
					}
					return true, nil
				}); err != nil {
					log.Println("walk SelectExprs failed:", err)
					return nil, fmt.Errorf("query is not supported:[%s]", query)
				}
			}
		}

		// record collected data
		log.Println("sql data:", collector)
		// handling unsupported case
		{
			if collector.IsAccessingVariable {
				return nil, fmt.Errorf("query is not supported:[%s]", query)
			}
		}
		// direct send
		if r, err := c.pgbackend.HandleSelect(query); err != nil {
			log.Println("handle select query failed:", err)
			return nil, fmt.Errorf("query is not supported:[%s]", query)
		} else {
			return r, nil
		}
	case *sqlparser.OtherRead:
		//The following operations are OtherRead
		//1. DESCRIBE
		//And not supported
		break
	default:
		break
	}

	log.Println(fmt.Errorf("query is not supported:[%s]", query))
	return nil, fmt.Errorf("query is not supported:[%s]", query)
}

func (c *commandHandler) HandleFieldList(table string, fieldWildcard string) ([]*mysql.Field, error) {
	log.Println(fmt.Errorf("field list is not supported:[%s], [%s]", table, fieldWildcard))
	return nil, fmt.Errorf("field list is not supported:[%s], [%s]", table, fieldWildcard)
}

func (c *commandHandler) HandleStmtPrepare(query string) (params int, columns int, context interface{}, err error) {
	log.Println(fmt.Errorf("stmt prepare is not supported:[%s]", query))
	return 0, 0, nil, fmt.Errorf("stmt prepare is not supported:[%s]", query)
}

func (c *commandHandler) HandleStmtExecute(context interface{}, query string, args []interface{}) (*mysql.Result, error) {
	log.Println(fmt.Errorf("stmt execute is not supported:[%s], [%s], [%s]", fmt.Sprint(context), query, fmt.Sprint(args)))
	return nil, fmt.Errorf("stmt execute is not supported:[%s], [%s], [%s]", fmt.Sprint(context), query, fmt.Sprint(args))
}

func (c *commandHandler) HandleStmtClose(context interface{}) error {
	log.Println(fmt.Errorf("stmt close is not supported:[%s]", fmt.Sprint(context)))
	return fmt.Errorf("stmt close is not supported:[%s]", fmt.Sprint(context))
}

func (c *commandHandler) HandleOtherCommand(cmd byte, data []byte) error {
	log.Println(fmt.Errorf("other command is not supported:[%d], [%s]", cmd, fmt.Sprint(data)))
	return fmt.Errorf("other command is not supported:[%d], [%s]", cmd, fmt.Sprint(data))
}

type MysqlIncomingHandler struct {
	listen    net.Listener
	pgbackend *PgBackend
}

func (m *MysqlIncomingHandler) Startup() error {
	go func() {
		for true {
			c, err := m.listen.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("incoming connection:", c.RemoteAddr())
			go func() {
				defer func(c net.Conn) {
					_ = c.Close()
				}(c)
				conn, err := server.NewConn(c, "root", "phpts", &commandHandler{
					AllowedDatabase: map[string]struct{}{
						"wordpress": {},
					},

					pgbackend: m.pgbackend,
				})
				if err != nil {
					log.Println(err)
					return
				}
				for {
					if err := conn.HandleCommand(); err != nil {
						log.Println(err)
						return
					}
				}
			}()
		}
	}()
	return nil
}

func NewMysqlIncomingHandler(addr string, pgbackend *PgBackend) *MysqlIncomingHandler {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	return &MysqlIncomingHandler{
		listen:    l,
		pgbackend: pgbackend,
	}
}
