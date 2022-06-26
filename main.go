package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	port           string = ":9010"
	dbname         string = "tododb"
	collectionname string = "todolist"
)

type (
	// DB Structure for todolist
	todoDbModel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Completed bool          `bson:"completed"`
		CreatedAt time.Time     `bson:"createAt"`
	}

	// UI Structure for todolist
	todoUIModel struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

// homeHandler - render homepage template file for todolist
func homeHandler(w http.ResponseWriter, r *http.Request) {

	err := render.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
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
			ID:        data.ID.Hex(),
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

	tododbData := todoDbModel{
		ID:        bson.NewObjectId(),
		Title:     todoUIData.Title,
		Completed: todoUIData.Completed,
		CreatedAt: time.Now(),
	}

	if err := db.C(collectionname).Insert(&tododbData); err != nil {
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

func deleteTodoList(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id) {
		render.JSON(w, http.StatusProcessing, renderer.M{
			"message": "This id is invalid",
		})
		return
	}

	if err := db.C(collectionname).RemoveId(bson.ObjectIdHex(id)); err != nil {
		render.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to delete todo",
			"error":   err,
		})
		return
	}

	render.JSON(w, http.StatusOK, renderer.M{
		"message": "todo delete sucessfully",
	})

}

func updatetodoList(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id) {
		render.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	tododata := todoUIModel{}

	if err := json.NewDecoder(r.Body).Decode(&tododata); err != nil {
		render.JSON(w, http.StatusProcessing, err)
		return
	}

	if tododata.Title == "" {
		render.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "the title field id required",
		})
		return
	}

	if err := db.C(collectionname).Update(
		bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"title": tododata.Title, "completed": tododata.Completed},
	); err != nil {
		render.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to update todo",
			"error":   err,
		})
		return
	}

	render.JSON(w, http.StatusOK, renderer.M{
		"message": "todolist Updated sucessfully",
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
	db = session.DB(dbname)
}

func todolistHandler() http.Handler {

	groupRouter := chi.NewRouter()
	groupRouter.Group(func(r chi.Router) {
		r.Get("/", getTodoList)
		r.Post("/", createTodoList)
		r.Put("/{id}", updatetodoList)
		r.Delete("/{id}", deleteTodoList)
	})
	return groupRouter
}

func main() {

	stopchan := make(chan os.Signal)
	signal.Notify(stopchan, os.Interrupt)

	homeRouter := chi.NewRouter()
	homeRouter.Use(middleware.Logger)

	homeRouter.Get("/", homeHandler)
	homeRouter.Mount("/todo", todolistHandler())

	server := &http.Server{
		Addr:         port,
		Handler:      homeRouter,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Print("Listen on port : ", port)
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
