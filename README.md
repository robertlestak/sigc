# sigc - signed client

sigc is a simple signed client proxy which enables a data source owner to precreate a query and sign it with a private key. The client can then provide a set of parameters and the signed query will be executed.

This allows untrusted clients to perform predefined transactions on a trusted data source.

## Configuration

A signed request is configured using a JSON object. The following options are available:

* `statement` - the data source statement to execute
* `param_count` - the number of parameters to expect
* `max_uses` - the maximum number of times the query can be executed. If this is set to 0, the query can be executed an unlimited number of times.
* `expires_at` - the time at which the query expires, in Unix timestamp format (seconds since epoc). If this is set to 0, the query never expires.
* `private_key` - the private key used to sign the query, base64 encoded.
* `connection` - the connection object for the data source
    * `driver` - the driver to use
    * `params` - a `map[string]string` of parameters for the driver. See the driver documentation for details.

## Usage

Create a signed transaction:

```bash
UNIX_TS_ONE_HOUR_FROM_NOW=$(($(date +%s) + 3600))
curl -X POST http://localhost:8080/sign \ 
    -d '{
        "statement": "INSERT INTO users (name, email) VALUES ($1, $2)",
        "param_count": 2,
        "max_uses": 1,
        "expires_at": '$UNIX_TS_ONE_HOUR_FROM_NOW',
        "private_key": "'$(cat private.pem | base64 -w0)'",
        "connection": {
            "driver": "postgres",
            "params": {
                "host": "localhost",
                "port": "5432",
                "user": "postgres",
                "pass": "postgres",
                "db": "postgres",
                "sslmode": "disable"
            }
        }
}'
```

The result will be a signed query which can be passed to the client:

```json
{"statement":"INSERT INTO users (name, email) VALUES ($1, $2)","param_count":2,"key_id":"c580d373-9e1a-4d6e-8d5e-ff00536c3345","signature":"b3885e251dc1ffcea0d1a6b396564a9deeeabd78654fec2c94c8f21a2dcbd09f44107d7762fcc1b08d129aa4f5eccc9f70f50011ceb9e7c02f7066c6a2885ff70a22bffcfd3945357c23b7fcccc8faffd373c20b54a4453d321722f51ce5e80c1229d2f466bfc03022c82664b91e79081da5da8c0a4370f86b56b7205fcaa3a9e90407b96f189722723c02eee2a81a8415eaf107ab250e4a23ea7a5aeb74ff677ba12b77aac6e9299d2fbfb64444bdd4a6aa3bb0a089542958134ff6cde4099aabf7af2b9cb181bc1c4223fdc6542230043b07d1ef2f6c606815d5c1f95f3973134ac6043a113733ae0ae0b7bc65d47d645c7a206eeb8b67498c9c13308ca910.23ade02b4129baddb7efe0f25a8cb2b9c4d10e159d4730585b3630795505d822db7c20419531b4d80b41483af9155e6d1a387388f9892429ca9266d62984f50b6013bafaaaad5f36e4eb40de8e20d6aaabb93d13c5c92ee7dc40dea5c3084f6fc0d8e0e52fd3343ae36295fe9a897b08607a110e7b7b0ed16d6f41edcc75fdd4be64122faefe94ad978e8af9420a122ca1a271d72c7d08596f4509ec4ca3bb0b1a24994d3a70b77b4d87d708a46eaded185aa36e5681dc054537a6789e8db447b87c6a1387df709b02723033b7931790226ff92e2566d0490e206f79d19f81ade11a00b749464d87a4d98add4ff018eb4008531f82095dd0804f3cb7fbe71440be2ea96c44e15ad5694422f96c32e5869d82bd94d6095d25abe31c72d876770a5ee7547d07ca85123475c4573cfd8ea3a5a0fce9b63ff5c38a60","expires_at":1668568903}
```

The client can then provide the parameters and execute the query:

```bash
curl -X POST http://localhost:8080/exec \
    -d '{"statement":"INSERT INTO users (name, email) VALUES ($1, $2)","param_count":2,"key_id":"c580d373-9e1a-4d6e-8d5e-ff00536c3345","signature":"b3885e251dc1ffcea0d1a6b396564a9deeeabd78654fec2c94c8f21a2dcbd09f44107d7762fcc1b08d129aa4f5eccc9f70f50011ceb9e7c02f7066c6a2885ff70a22bffcfd3945357c23b7fcccc8faffd373c20b54a4453d321722f51ce5e80c1229d2f466bfc03022c82664b91e79081da5da8c0a4370f86b56b7205fcaa3a9e90407b96f189722723c02eee2a81a8415eaf107ab250e4a23ea7a5aeb74ff677ba12b77aac6e9299d2fbfb64444bdd4a6aa3bb0a089542958134ff6cde4099aabf7af2b9cb181bc1c4223fdc6542230043b07d1ef2f6c606815d5c1f95f3973134ac6043a113733ae0ae0b7bc65d47d645c7a206eeb8b67498c9c13308ca910.23ade02b4129baddb7efe0f25a8cb2b9c4d10e159d4730585b3630795505d822db7c20419531b4d80b41483af9155e6d1a387388f9892429ca9266d62984f50b6013bafaaaad5f36e4eb40de8e20d6aaabb93d13c5c92ee7dc40dea5c3084f6fc0d8e0e52fd3343ae36295fe9a897b08607a110e7b7b0ed16d6f41edcc75fdd4be64122faefe94ad978e8af9420a122ca1a271d72c7d08596f4509ec4ca3bb0b1a24994d3a70b77b4d87d708a46eaded185aa36e5681dc054537a6789e8db447b87c6a1387df709b02723033b7931790226ff92e2566d0490e206f79d19f81ade11a00b749464d87a4d98add4ff018eb4008531f82095dd0804f3cb7fbe71440be2ea96c44e15ad5694422f96c32e5869d82bd94d6095d25abe31c72d876770a5ee7547d07ca85123475c4573cfd8ea3a5a0fce9b63ff5c38a60","expires_at":1668568903, "params": ["John Doe", "example@example.com"]}'
```

The server will then verify the signature and execute the query. If there are any results, they will be returned as a JSON array, and errors are returned if there are any.


```json
{"results":null,"error":{"Severity":"ERROR","Code":"42703","Message":"column \"name\" of relation \"users\" does not exist","Detail":"","Hint":"","Position":"20","InternalPosition":"","InternalQuery":"","Where":"","Schema":"","Table":"","Column":"","DataTypeName":"","Constraint":"","File":"parse_target.c","Line":"1061","Routine":"checkInsertTargets"}}
```

## Drivers

The following drivers are currently available:

- [cassandra](#cassandra)
- [cockroachdb]()
- [mssql]()
- [mysql]()
- [postgres]()
- [scylla]()

Below are the parameters that are available for each driver.

### Cassandra

* `hosts` - A comma-separated list of Cassandra hosts to connect to.
* `user` - The username to use when connecting to Cassandra.
* `pass` - The password to use when connecting to Cassandra.
* `keyspace` - The keyspace to use when connecting to Cassandra.
* `consistency` - The consistency level to use when connecting to Cassandra.

### CockroachDB

* `host` - The host to connect to.
* `port` - The port to connect to.
* `user` - The username to use when connecting to CockroachDB.
* `pass` - The password to use when connecting to CockroachDB.
* `db` - The database to use when connecting to CockroachDB.
* `sslmode` - The SSL mode to use when connecting to CockroachDB.
* `sslrootcert` - The SSL root certificate to use when connecting to CockroachDB.
* `sslcert` - The SSL certificate to use when connecting to CockroachDB.
* `sslkey` - The SSL key to use when connecting to CockroachDB.
* `routing_id` - The routing ID to use when connecting to CockroachDB.

### MSSQL

* `host` - The host to connect to.
* `port` - The port to connect to.
* `user` - The username to use when connecting to MSSQL.
* `pass` - The password to use when connecting to MSSQL.
* `db` - The database to use when connecting to MSSQL.

### MySQL

* `host` - The host to connect to.
* `port` - The port to connect to.
* `user` - The username to use when connecting to MySQL.
* `pass` - The password to use when connecting to MySQL.
* `db` - The database to use when connecting to MySQL.

### Postgres

* `host` - The host to connect to.
* `port` - The port to connect to.
* `user` - The username to use when connecting to Postgres.
* `pass` - The password to use when connecting to Postgres.
* `db` - The database to use when connecting to Postgres.
* `sslmode` - The SSL mode to use when connecting to Postgres.
* `sslrootcert` - The SSL root certificate to use when connecting to Postgres.
* `sslcert` - The SSL certificate to use when connecting to Postgres.
* `sslkey` - The SSL key to use when connecting to Postgres.

### Scylla

* `hosts` - A comma-separated list of Scylla hosts to connect to.
* `user` - The username to use when connecting to Scylla.
* `pass` - The password to use when connecting to Scylla.
* `keyspace` - The keyspace to use when connecting to Scylla.
* `consistency` - The consistency level to use when connecting to Scylla.
* `local_dc` - The local datacenter to use when connecting to Scylla.