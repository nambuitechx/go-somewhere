package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/table"
)

type Log struct {
	ID			gocql.UUID
	Text		string
	CreatedAt	time.Time
}

var logMetadata = table.Metadata{
	Name: "logs",
	Columns: []string{"id", "text", "created_at"},
	PartKey: []string{"id"},
	SortKey: []string{"created_at"},
}

var logTable = table.New(logMetadata)

func main()  {
	hosts := []string{"localhost:9042"}

	// Create gocql cluster.
	cluster := gocql.NewCluster(hosts...)

	// Wrap session on creation, gocqlx session embeds gocql.Session pointer.
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	log.Println("Connected to Scylla successfully")

	for i := 0; i < 10; i++ {
		l := Log{
			ID: gocql.TimeUUID(),
			Text: strconv.Itoa(i),
			CreatedAt: time.Now(),
		}
		q := session.Query(logTable.Insert()).BindStruct(l)
		if err := q.ExecRelease(); err != nil {
			log.Printf("Failed to insert log to scylla: %v", err.Error())
			continue
		}
	}
}
