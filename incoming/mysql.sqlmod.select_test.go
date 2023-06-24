package incoming

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestModifyLimit(t *testing.T) {
	sql := ` SELECT DISTINCT t.term_id, tr.object_id
                        FROM wp_terms AS t  INNER JOIN wp_term_taxonomy AS tt ON t.term_id = tt.term_id INNER JOIN wp_term_relationships AS tr ON tr.term_taxonomy_id = tt.term_taxonomy_id
                        WHERE tt.taxonomy IN ('wp_theme') AND tr.object_id IN (0)
                        ORDER BY t.name ASC`

	t.Log(sql)
}

func doNewListenMysql(addr string) (*MysqlIncomingHandler, error) {
	pgbackend, err := NewPgBackend("postgres://admin:admin@localhost:5432/wordpress")
	if err != nil {
		return nil, err
	}
	in := NewMysqlIncomingHandler(addr, pgbackend)
	return in, in.Startup()
}

type Post struct {
	ID                  int
	Author              int
	PostDate            time.Time
	PostDateGMT         time.Time
	PostContent         string
	PostTitle           string
	PostExcerpt         string
	PostStatus          string
	CommentStatus       string
	PingStatus          string
	PostsPassword       string
	PostName            string
	ToPing              string
	Pinged              string
	PostModified        string
	PostModifiedGMT     string
	PostContentFiltered string
	PostParent          int
	Guid                string
	MenuOrder           int
	PostType            string
	PostMimeType        string
	CommentCount        int
}

func TestSample(t *testing.T) {
	h, err := doNewListenMysql("127.0.0.1:33306")
	if err != nil {
		t.Fatal(err)
	}
	var _ = h

	db, err := sql.Open("mysql", "root:phpts@tcp(127.0.0.1:33306)/wordpress?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from wp_posts")
	if err != nil {
		t.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		obj := Post{}

		t.Log("---->>")
		ts, err := rows.ColumnTypes()
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range ts {
			t.Log(v.Name(), v.DatabaseTypeName())
		}
		if err := rows.Scan(&obj.ID, &obj.Author, &obj.PostDate, &obj.PostDateGMT, &obj.PostContent,
			&obj.PostTitle, &obj.PostExcerpt, &obj.PostStatus, &obj.CommentStatus, &obj.PingStatus,
			&obj.PostsPassword, &obj.PostName, &obj.ToPing, &obj.Pinged, &obj.PostModified, &obj.PostModifiedGMT,
			&obj.PostContentFiltered, &obj.PostParent, &obj.Guid, &obj.MenuOrder, &obj.PostType, &obj.PostMimeType, &obj.CommentCount); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(data))
		t.Log("<<----")
	}

	fmt.Println(111)
}
