package main

import (
	"log"

	"github.com/thedevsaddam/renderer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var render *renderer.Render
var db *mgo.Database

const (
	hostname       string = "localhost:27017"
	port           string = ":9000"
	dbname         string = "tododb"
	collectionname string = "todo"
)

type (
	todoDbModel struct {
		ID        bson.ObjectId `bson: _id, omitempty`
		Title     string        `bson: "title"`
		Completed bool          `bson : "complete"`
		CreatedAt string        `bson : "createdat"`
	}

	todoUIModel struct {
		ID        bson.ObjectId `json: "id"`
		Title     string        `json: "title"`
		Completed bool          `json: "completed"`
		CreatedAt string        `json: createdat`
	}
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	render = renderer.New()
	session, err := mgo.Dial(hostname)
	checkError(err)
	session.SetMode(mgo.Monotonic, true)
	session.DB(dbname)
}

func main() {

}
