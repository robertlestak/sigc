package utils

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

func RowsToMap(rows *sql.Rows) (map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "RowsToMap",
	})
	l.Debug("Converting row to map")
	m := make(map[string]interface{})
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			l.Error(err)
			return nil, err
		}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			l.Error(err)
			return nil, err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
			l.Debugf("%s: %s", colName, *val)
		}
	}
	l.Debug("Converted row to map")
	return m, nil
}

func RowsToMapSlice(rows *sql.Rows) ([]map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "RowsToMap",
	})
	l.Debug("Converting row to map")
	var sm []map[string]any
	for rows.Next() {
		m := make(map[string]interface{})
		cols, err := rows.Columns()
		if err != nil {
			l.Error(err)
			return nil, err
		}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			l.Error(err)
			return nil, err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
			l.Debugf("%s: %s", colName, *val)
		}
		sm = append(sm, m)
	}
	l.Debug("Converted row to map")
	return sm, nil
}
