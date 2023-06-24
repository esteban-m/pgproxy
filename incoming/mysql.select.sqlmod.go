package incoming

import (
	"errors"
	"fmt"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

var selectSQLModifications SelectSQLModifications

func init() {
	selectSQLModifications = []SelectSQLModification{
		SelectSQLModificationTool{}.LimitAndOffset,

		SelectSQLModificationTool{}.FixDistinctOrderByFieldShouldBeInSelectExpr,
		SelectSQLModificationTool{}.AddQuotesToSupportIdentifierCaseSensitive,
	}
}

type ModificationSpec struct {
	AddedLastMetaColumnCnt int
}

type SelectSQLModifications []SelectSQLModification

func (m SelectSQLModifications) Run(in string) (string, ModificationSpec, error) {
	modSpec := ModificationSpec{}
	for _, v := range m {
		if out, mod, err := v(in); err != nil {
			return "", modSpec, err
		} else {
			in = out
			if mod.AddedLastMetaColumnCnt > 0 {
				modSpec.AddedLastMetaColumnCnt += mod.AddedLastMetaColumnCnt
			}
		}
	}
	return in, modSpec, nil
}

type SelectSQLModification func(in string) (string, ModificationSpec, error)

type SelectSQLModificationTool struct {
}

func (s SelectSQLModificationTool) LimitAndOffset(in string) (string, ModificationSpec, error) {
	st, err := sqlparser.Parse(in)
	if err != nil {
		return "", ModificationSpec{}, err
	}

	selectStatement, ok := st.(*sqlparser.Select)
	if !ok {
		return "", ModificationSpec{}, errors.New("not a select query")
	}
	newSelectStatement := &ModifiedSelect{
		Cache:       selectStatement.Cache,
		Comments:    selectStatement.Comments,
		Distinct:    selectStatement.Distinct,
		Hints:       selectStatement.Hints,
		SelectExprs: selectStatement.SelectExprs,
		From:        selectStatement.From,
		Where:       selectStatement.Where,
		GroupBy:     selectStatement.GroupBy,
		Having:      selectStatement.Having,
		OrderBy:     selectStatement.OrderBy,
		Lock:        selectStatement.Lock,
	}
	if selectStatement.Limit != nil {
		newSelectStatement.Limit = &LimitOffset{*selectStatement.Limit}
	}

	buf := sqlparser.NewTrackedBuffer(nil)
	newSelectStatement.Format(buf)
	pq := buf.ParsedQuery()

	return pq.Query, ModificationSpec{}, nil
}

func (s SelectSQLModificationTool) AddQuotesToSupportIdentifierCaseSensitive(in string) (string, ModificationSpec, error) {
	// mysql may leverage case sensitive identifier depending on operating system
	// some php systems like wordpress may require exactly the same case of identifier for ORM
	// By default, in postgresql, if identifiers are not quoted, they will be automatically converted to lower case.
	// So quotes are required to enforce case sensitive in postgresql
	// Note: * should not be quoted.
	// Note2: constant where clause should be prevented, e.g. 1=1 .

	st, err := sqlparser.Parse(in)
	if err != nil {
		return "", ModificationSpec{}, err
	}

	selectStatement, ok := st.(*sqlparser.Select)
	if !ok {
		return "", ModificationSpec{}, errors.New("not a select query")
	}

	fmt.Println("----------------")
	if selectStatement.SelectExprs != nil {
		where := selectStatement.Where
		switch exp := where.Expr.(type) {
		case *sqlparser.ComparisonExpr:
			if colName, ok := exp.Left.(*sqlparser.ColName); ok {
				//FIXME automatically add backquote by sqlparser
				colName.Name = sqlparser.NewColIdent(fmt.Sprint(`"`, colName.Name.String(), `"`))
				exp.Left = colName
			}
		}
		selectStatement.Where = where
	}

	buf := sqlparser.NewTrackedBuffer(nil)
	selectStatement.Format(buf)
	pq := buf.ParsedQuery()
	fmt.Println(pq.Query)
	fmt.Println("----------------")

	return in, ModificationSpec{}, nil
}

func (s SelectSQLModificationTool) FixDistinctOrderByFieldShouldBeInSelectExpr(in string) (string, ModificationSpec, error) {
	// mysql allows distinct+orderby fields are not in select expression list
	// but postgresql doesn't allow this

	//FIXME this solution have to be well considered since the meaning of the sql may have been changed.

	// Cases:
	// 1. distinct + order by and order by field doesn't in select expression
	// 2. select expression may have several alias tables and in order to avoid name mapping failed, same field name have to be retained
	// 3. select expression may have wildcard(*)

	// solution
	// 1. add one or certain count of columns in the select expression for distinct(question: the same behavior?)
	// 2. increment AddedLastMetaColumnCnt for ResultSet to exclude those fields

	st, err := sqlparser.Parse(in)
	if err != nil {
		return "", ModificationSpec{}, err
	}

	selectStatement, ok := st.(*sqlparser.Select)
	if !ok {
		return "", ModificationSpec{}, errors.New("not a select query")
	}

	fmt.Println("================")
	fmt.Println(selectStatement.SelectExprs)
	fmt.Println(selectStatement.Distinct)
	fmt.Println(selectStatement.OrderBy)
	fmt.Println("================")

	//TODO
	return in, ModificationSpec{}, nil
}

type ModifiedSelect struct {
	Cache       string
	Comments    sqlparser.Comments
	Distinct    string
	Hints       string
	SelectExprs sqlparser.SelectExprs
	From        sqlparser.TableExprs
	Where       *sqlparser.Where
	GroupBy     sqlparser.GroupBy
	Having      *sqlparser.Where
	OrderBy     sqlparser.OrderBy
	Limit       *LimitOffset
	Lock        string
}

func (node *ModifiedSelect) Format(buf *sqlparser.TrackedBuffer) {
	buf.Myprintf("select %v%s%s%s%v from %v%v%v%v%v%v%s",
		node.Comments, node.Cache, node.Distinct, node.Hints, node.SelectExprs,
		node.From, node.Where,
		node.GroupBy, node.Having, node.OrderBy,
		node.Limit, node.Lock)
}

type LimitOffset struct {
	sqlparser.Limit
}

func (node *LimitOffset) Format(buf *sqlparser.TrackedBuffer) {
	if node == nil {
		return
	}
	buf.Myprintf(" limit ")
	buf.Myprintf("%v", node.Rowcount)
	if node.Offset != nil {
		buf.Myprintf(" offset %v", node.Offset)
	}
}
