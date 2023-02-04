package translator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

type DIYSqlHandler struct {
	DefaultSqlHandler
}

func (*DIYSqlHandler) NeedHandleCreate(sql string, stat *sqlparser.CreateTable) (bool, error) {
	// looks like all options should be replaced
	if len(stat.Options) > 0 {
		return true, nil
	}
	// auto increment
	for _, v := range stat.Columns {
		if len(v.Options) > 0 {
			for _, v2 := range v.Options {
				if v2.Type == sqlparser.ColumnOptionAutoIncrement {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (*DIYSqlHandler) HandleCreate(sql string, stat *sqlparser.CreateTable) (string, error) {
	// modify ast
	//TODO cleanup ast
	stat.Options = nil

	// generate new sql
	buf := sqlparser.NewTrackedBuffer(nil)
	stat.Format(buf)
	pq := buf.ParsedQuery()
	return pq.Query, nil
}

func TestExampleCreate(t *testing.T) {
	handler := new(DIYSqlHandler)
	req, err := ParseSQL(`
CREATE TABLE wp_commentmeta (
    meta_id bigint(20) unsigned NOT NULL auto_increment,
    comment_id bigint(20) unsigned NOT NULL default '0',
    meta_key varchar(255) default NULL,
    meta_value longtext,
    PRIMARY KEY  (meta_id),
    KEY comment_id (comment_id),
    KEY meta_key (meta_key(191))
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci
`, handler)
	t.Log(req, err)
}

func TestExampleDelete(t *testing.T) {
	stmt, err := sqlparser.Parse("DELETE FROM `wp_options` WHERE `option_name` = '_transient_global_styles_svg_filters_twentytwentythree'")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Delete); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleDescribe(t *testing.T) {
	stmt, err := sqlparser.Parse("DESCRIBE wp_comments;")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.OtherRead); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleInsert(t *testing.T) {
	stmt, err := sqlparser.Parse("INSERT IGNORE INTO `wp_options` ( `option_name`, `option_value`, `autoload` ) VALUES ('auto_updater.lock', '1674463706', 'no') /* LOCK */")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Insert); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleSelect(t *testing.T) {
	stmt, err := sqlparser.Parse("SELECT * FROM wp_posts  WHERE (post_type = 'page' AND post_status = 'publish')     ORDER BY menu_order,wp_posts.post_title ASC")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Select); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleSelect2(t *testing.T) {
	stmt, err := sqlparser.Parse("SELECT 1 as test FROM wp_posts WHERE post_type = 'post' AND post_status = 'publish' LIMIT 1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Select); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleSelect3(t *testing.T) {
	stmt, err := sqlparser.Parse("SELECT @@SESSION.sql_mode")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Select); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

// FIXME failed!!!!!!!!!!!!
func TestExampleSet(t *testing.T) {
	stmt, err := sqlparser.Parse("SET NAMES 'utf8mb4' COLLATE 'utf8mb4_unicode_520_ci'")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))
	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleShow(t *testing.T) {
	stmt, err := sqlparser.Parse("SHOW FULL COLUMNS FROM `wp_postmeta`")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Show); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}

func TestExampleUpdate(t *testing.T) {
	stmt, err := sqlparser.Parse("UPDATE `wp_term_taxonomy` SET `count` = 1 WHERE `term_taxonomy_id` = 1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stmt = %+v\n", stmt)
	fmt.Printf("stmt = %+v\n", reflect.TypeOf(stmt))

	if s, ok := stmt.(*sqlparser.Update); ok {
		fmt.Println("TARGET TYPE:", s)
	} else {
		t.Fatal("undesired type")
	}

	stmt.WalkSubtree(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		if node == nil {
			fmt.Println("node=", node)
		} else {
			fmt.Println("node=", reflect.TypeOf(node))
		}
		return true, nil
	})
}
