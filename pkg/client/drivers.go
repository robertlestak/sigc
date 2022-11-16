package client

import (
	"github.com/robertlestak/sigc/drivers/cassandra"
	"github.com/robertlestak/sigc/drivers/cockroachdb"
	"github.com/robertlestak/sigc/drivers/mssql"
	"github.com/robertlestak/sigc/drivers/mysql"
	"github.com/robertlestak/sigc/drivers/postgres"
	"github.com/robertlestak/sigc/drivers/scylla"
)

type DriverName string

var (
	DriverCassandra   DriverName = "cassandra"
	DriverCockroachDB DriverName = "cockroachdb"
	DriverPostgres    DriverName = "postgres"
	DriverMSsql       DriverName = "mssql"
	DriverMysql       DriverName = "mysql"
	DriverScylla      DriverName = "scylla"
)

func GetDriver(driver DriverName) Client {
	switch driver {
	case DriverCassandra:
		return &cassandra.Cassandra{}
	case DriverCockroachDB:
		return &cockroachdb.CockroachDB{}
	case DriverPostgres:
		return &postgres.Postgres{}
	case DriverMSsql:
		return &mssql.MSSql{}
	case DriverMysql:
		return &mysql.Mysql{}
	case DriverScylla:
		return &scylla.Scylla{}
	}
	return nil
}
