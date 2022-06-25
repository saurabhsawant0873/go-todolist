package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	// DB Structure for todolist
	todoDbModel struct {
		ID        bson.ObjectId `bson: _id, omitempty`
		Title     string        `bson: "title"`
		Completed bool          `bson : "complete"`
		CreatedAt time.Time     `bson : "createdat"`
	}

	// UI Structure for todolist
	todoUIModel struct {
		ID        bson.ObjectId `json: "id"`
		Title     string        `json: "title"`
		Completed bool          `json: "completed"`
		CreatedAt time.Time     `json: createdat`
	}
)

// homeHandler - render homepage template file for todolist
func homeHandler(w http.ResponseWriter, r *http.Request) {

	err := render.Template(w, http.StatusProcessing, []string{"static/home.tpl"}, nil)
	checkError(err)
}

func getTodoList(w http.ResponseWriter, r *http.Request) {
	todoData := []todoDbModel{}

	if err := db.C(collectionname).Find(bson.M{}).All(&todoData); err != nil {
		render.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to fetch todo",
			"error":   err,
		})
		return
	}

	todoDataToUI := []todoUIModel{}

	for _, data := range todoData {
		todoDataToUI = append(todoDataToUI, todoUIModel{
			ID:        data.ID,
			Title:     data.Title,
			Completed: data.Completed,
			CreatedAt: data.CreatedAt,
		})
	}

	render.JSON(w, http.StatusOK, renderer.M{
		"data": todoDataToUI,
	})
}

func createTodoList(w http.ResponseWriter, r *http.Request) {

	todoUIData := todoUIModel{}

	if err := json.NewDecoder(r.Body).Decode(&todoUIData); err != nil {
		render.JSON(w, http.StatusBadRequest, err)
		return
	}

	tododbData := &todoDbModel{
		ID:        todoUIData.ID,
		Title:     todoUIData.Title,
		Completed: todoUIData.Completed,
		CreatedAt: time.Now(),
	}

	if err := db.C(collectionname).Insert(tododbData); err != nil {
		render.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to save to do",
			"error":   err,
		})
		return
	}

	render.JSON(w, http.StatusCreated, renderer.M{
		"message": "todo created sucessfully",
		"todo_id": tododbData.ID.Hex(),
	})
}

// checkError - Check and print if any errors
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	render = renderer.New()
	session, err := mgo.Dial(hostname)
	checkError(err)
	// Three modes: Strong -> More stability with read and write operations
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

	stopchan := make(chan os.Signal)
	signal.Notify(stopchan, os.Interrupt)

	homeRouter := chi.NewRouter()
	homeRouter.Use(middleware.Logger)

	homeRouter.Get("/", homeHandler)
	homeRouter.Mount("/todo", todolistHandler)

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

	<-stopchan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)

	defer cancel()
	log.Println("Shutdown server gracefully")
}
