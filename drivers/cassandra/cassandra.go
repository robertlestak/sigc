package cassandra

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/robertlestak/sigc/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type Cassandra struct {
	Client      *gocql.Session
	Hosts       []string
	User        string
	Password    string
	Consistency string
	Keyspace    string
}

func (d *Cassandra) parseParams(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "cassandra",
		"fn":  "parseParams",
	})
	l.Debug("start")
	if params["hosts"] == nil {
		return fmt.Errorf("hosts is required")
	}
	d.Hosts = strings.Split(params["hosts"].(string), ",")
	if params["user"] == nil {
		return fmt.Errorf("user is required")
	}
	d.User = params["user"].(string)
	if params["pass"] == nil {
		return fmt.Errorf("pass is required")
	}
	d.Password = params["pass"].(string)
	if params["consistency"] == nil {
		return fmt.Errorf("consistency is required")
	}
	d.Consistency = params["consistency"].(string)
	if params["keyspace"] == nil {
		return fmt.Errorf("keyspace is required")
	}
	d.Keyspace = params["keyspace"].(string)
	return nil
}

func (d *Cassandra) Connect(params map[string]any) error {
	l := log.WithFields(log.Fields{
		"app": "cassandra",
		"fn":  "Connect",
	})
	l.Debug("start")
	if err := d.parseParams(params); err != nil {
		return err
	}
	cluster := gocql.NewCluster(d.Hosts...)
	// parse consistency string
	consistencyLevel := gocql.ParseConsistency(d.Consistency)
	cluster.Consistency = consistencyLevel
	if d.Keyspace != "" {
		cluster.Keyspace = d.Keyspace
	}
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	if d.User != "" || d.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: d.User, Password: d.Password}
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	d.Client = session
	return nil
}

func (d *Cassandra) Disconnect() error {
	l := log.WithFields(log.Fields{
		"app": "cassandra",
		"fn":  "Disconnect",
	})
	l.Debug("start")
	d.Client.Close()
	l.Debug("Disconnected")
	return nil
}

func (d *Cassandra) Exec(r *schema.Request) *schema.Response {
	l := log.WithFields(log.Fields{
		"app": "cassandra",
		"fn":  "Exec",
	})
	l.Debug("start")
	var err error
	res := &schema.Response{}
	l.Debug("Executing statement: ", r.Statement)
	qry := d.Client.Query(r.Statement, r.Params...)
	defer qry.Release()
	m, err := schema.CqlRowsToMapSlice(qry)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil
		}
		return &schema.Response{
			Error: err,
		}
	}
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
