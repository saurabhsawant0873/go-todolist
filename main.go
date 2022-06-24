package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
	// Three modes: Strong -> More stability withread and write operations
	// Monotonic -> Somewhere between Strong and Eventual
	// Eventual -> Focusses more on loadbalancing
	session.SetMode(mgo.Monotonic, true)
	session.DB(dbname)
}

func todolistHandler() http.Handler {

	groupRouter := chi.NewRouter()
	groupRouter.Group(func(r chi.Router) {
		groupRouter.Get("/", getTodoList)
		groupRouter.Post("/", createTodoList)
		groupRouter.Put("/{id}", updatetodoList)
		groupRouter.Delete("/{id}", deleteTodoList)
	})
	return groupRouter
}

func main() {

	homeRouter := chi.NewRouter()
	homeRouter.Use(middleware.Logger)

	homeRouter.Get("/", homeHandler)
	homeRouter.Mount("/todolist", todolistHandler)

	server := &http.Server{
		Addr:         port,
		Handler:      homeRouter,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Listen on port : %d", port)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Listen : %s \n", err)
		}
	}()
}
