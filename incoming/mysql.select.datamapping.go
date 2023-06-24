package incoming

import (
	"time"
)

var selectResultSetDataMappings SelectResultSetDataMappings

func init() {
	selectResultSetDataMappings = SelectResultSetDataMappings{
		SelectResultSetDataMapping{}.timeMapping,
	}
}

type SelectResultSetDataMappings []func([][]interface{}) [][]interface{}

func (s SelectResultSetDataMappings) Run(in [][]interface{}) ([][]interface{}, error) {
	for _, v := range s {
		in = v(in)
	}
	return in, nil
}

type SelectResultSetDataMapping struct {
}

func (s SelectResultSetDataMapping) timeMapping(in [][]interface{}) [][]interface{} {
	for idx1, v := range in {
		for idx2, vv := range v {
			if tt, ok := vv.(time.Time); ok {
				v[idx2] = tt // use original value now since time.Time support is builtin conversion function
				//fmt.Println("hit time.Time [", v[idx2], "]")
			}
		}
		in[idx1] = v
	}
	return in
}
