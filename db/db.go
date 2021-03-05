package db

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type DB struct {
	name string
}

func (db *DB) connect() {
	influxdb2.NewClient("http://localhost:8086", "my-token")
}
