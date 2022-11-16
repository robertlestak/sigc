package mssql

import (
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/robertlestak/sigc/internal/utils"
	"github.com/robertlestak/sigc/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type MSSql struct {
	Client *sql.DB
	Host   string
	Port   string
	User   string
	Pass   string
	Db     string
}

func (d *MSSql) parseParams(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "mysql",
		"fn":  "parseParams",
	})
	l.Debug("start")
	if params["host"] == nil {
		return fmt.Errorf("host is required")
	}
	d.Host = params["host"].(string)
	if params["port"] == nil {
		return fmt.Errorf("port is required")
	}
	d.Port = params["port"].(string)
	if params["user"] == nil {
		return fmt.Errorf("user is required")
	}
	d.User = params["user"].(string)
	if params["pass"] == nil {
		return fmt.Errorf("pass is required")
	}
	d.Pass = params["pass"].(string)
	if params["db"] == nil {
		return fmt.Errorf("db is required")
	}
	d.Db = params["db"].(string)
	return nil
}

func (d *MSSql) Connect(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "mysql",
		"fn":  "Connect",
	})
	l.Debug("start")
	if err := d.parseParams(params); err != nil {
		return err
	}
	var err error
	connStr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", d.Host, d.User, d.Pass, d.Port, d.Db)
	l.Debug("Connecting to mssql: ", connStr)
	d.Client, err = sql.Open("mssql", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Initialized mssql client")
	// ping the database to check if it is alive
	err = d.Client.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Connected")
	return nil
}

func (d *MSSql) Disconnect() error {
	l := log.WithFields(log.Fields{
		"app": "mysql",
		"fn":  "Disconnect",
	})
	l.Debug("start")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Disconnected")
	return nil
}

func (d *MSSql) Exec(r *schema.Request) *schema.Response {
	l := log.WithFields(log.Fields{
		"app": "mysql",
		"fn":  "Exec",
	})
	l.Debug("start")
	var err error
	res := &schema.Response{}
	l.Debug("Executing statement: ", r.Statement)
	resp, err := d.Client.Query(r.Statement, r.Params...)
	if err != nil {
		l.Error(err)
		return &schema.Response{
			Results: nil,
			Error:   err,
		}
	}
	defer resp.Close()
	m, err := utils.RowsToMapSlice(resp)
	if err != nil {
		l.Error(err)
		return &schema.Response{
			Results: nil,
			Error:   err,
		}
	}
	res.Results = m
	return res
}
