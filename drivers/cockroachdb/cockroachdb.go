package cockroachdb

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/robertlestak/sigc/internal/utils"
	"github.com/robertlestak/sigc/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type CockroachDB struct {
	Client      *sql.DB
	Host        string
	Port        string
	User        string
	Pass        string
	Db          string
	SslMode     string
	SSLRootCert *string
	SSLCert     *string
	SSLKey      *string
	RoutingID   *string
}

func (d *CockroachDB) parseParams(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "cockroachdb",
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
	if params["sslmode"] == nil {
		return fmt.Errorf("sslmode is required")
	}
	d.SslMode = params["sslmode"].(string)
	if params["sslrootcert"] != nil {
		d.SSLRootCert = params["sslrootcert"].(*string)
	}
	if params["sslcert"] != nil {
		d.SSLCert = params["sslcert"].(*string)
	}
	if params["sslkey"] != nil {
		d.SSLKey = params["sslkey"].(*string)
	}
	if params["routing_id"] != nil {
		d.RoutingID = params["routing_id"].(*string)
	}
	return nil
}

func (d *CockroachDB) Connect(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "cockroachdb",
		"fn":  "Connect",
	})
	l.Debug("start")
	if err := d.parseParams(params); err != nil {
		return err
	}
	var err error
	var opts string
	var connStr string = "cockroachdbql://"
	if d.RoutingID != nil && *d.RoutingID != "" {
		opts = "&options=--cluster%3D" + *d.RoutingID
	}
	if d.User != "" && d.Pass != "" {
		connStr += fmt.Sprintf("%s:%s@%s:%s/%s",
			d.User, d.Pass, d.Host, d.Port, d.Db)
	} else if d.User != "" && d.Pass == "" {
		connStr += fmt.Sprintf("%s@%s:%s/%s",
			d.User, d.Host, d.Port, d.Db)
	}
	connStr += "?sslmode=" + d.SslMode
	if d.SSLRootCert != nil && *d.SSLRootCert != "" {
		connStr += "&sslrootcert=" + *d.SSLRootCert
	}
	if d.SSLCert != nil && *d.SSLCert != "" {
		connStr += "&sslcert=" + *d.SSLCert
	}
	if d.SSLKey != nil && *d.SSLKey != "" {
		connStr += "&sslkey=" + *d.SSLKey
	}
	connStr += opts
	l.Debugf("Connecting to %s", connStr)
	d.Client, err = sql.Open("cockroachdb", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	err = d.Client.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Connected")
	return nil
}

func (d *CockroachDB) Disconnect() error {
	l := log.WithFields(log.Fields{
		"app": "cockroachdb",
		"fn":  "Disconnect",
	})
	l.Debug("start")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Disconnected from psql")
	return nil
}

func (d *CockroachDB) Exec(r *schema.Request) *schema.Response {
	l := log.WithFields(log.Fields{
		"app": "cockroachdb",
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
