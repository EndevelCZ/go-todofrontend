package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/EndevelCZ/todo/pb"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// type frontendServer struct {
// 	storageSvcAddr string
// 	storageSvcConn *grpc.ClientConn
// }

const (
	TodoSvcAddr = "127.0.0.1:5000"
)

// var (
// 	templates = template.Must(template.New("").
// 		Funcs(template.FuncMap{
// 			"renderMoney": renderMoney,
// 		}).ParseGlob("templates/*.html"))
// )

func main() {
	r := mux.NewRouter()
	files := http.FileServer(http.Dir("./public"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", files))
	// r.HandleFunc("/todos", TodoHandler)
	r.HandleFunc("/todos", TodoHandler).Methods("GET")
	r.HandleFunc("/posttodos", PostTodoHtmlHandler)
	r.HandleFunc("/todos/{id}", DeleteUserHandler).Methods("DELETE")
	r.HandleFunc("/todos/{id}", UpdateUserHandler).Methods("PATCH")

	// r.HandleFunc("/todos/{category}", TodoCategoryHandler)
	// r.HandleFunc("/todos/{category}/{id:[0-9]+}", TodoCategoriesHandler)
	// r.HandleFunc("/storage", StorageHandler)
	// r.HandleFunc("/", TodoHandler)
	http.Handle("/", r)

	addr := "127.0.0.1:8000"
	fmt.Println(addr)
	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("listening on: %s", addr)
	log.Fatal(srv.ListenAndServe())
}
func renderHTTPError(r *http.Request, w http.ResponseWriter, err error, code int) {

	errMsg := fmt.Sprintf("%+v", err)

	w.WriteHeader(code)
	fmt.Fprintf(w, errMsg)
	// templates.ExecuteTemplate(w, "error", map[string]interface{}{
	// 	"session_id":  sessionID(r),
	// 	"request_id":  r.Context().Value(ctxKeyRequestID{}),
	// 	"error":       errMsg,
	// 	"status_code": code,
	// 	"status":      http.StatusText(code)})
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	fmt.Printf("TODO patching id %d", id)
	var t pb.Todo
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&t)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}

	ctx := context.Background()
	// conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx,
		TodoSvcAddr,
		grpc.WithInsecure(),
		// grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
	)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", TodoSvcAddr))
	}

	defer conn.Close()
	client := pb.NewTodosClient(conn)

	tid := &pb.Integer{
		Id: id,
	}
	todo, err := client.CheckTodo(ctx, tid)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	logrus.Printf("%#v", todo)
	b, err := json.Marshal(todo)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	fmt.Fprintf(w, string(b))
	log.Println("update:", time.Since(start))

}
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method == "DELETE" {
	// 	params := mux.Vars(r)
	// }
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	fmt.Printf("TODO deleting id %s", id)

	ctx := context.Background()
	// conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx,
		TodoSvcAddr,
		grpc.WithInsecure(),
		// grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
	)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", TodoSvcAddr))
	}

	defer conn.Close()
	client := pb.NewTodosClient(conn)

	tid := &pb.Integer{
		Id: id,
	}
	todo, err := client.DeleteTodo(ctx, tid)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	logrus.Printf("%#v", todo)
	b, err := json.Marshal(todo)
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	fmt.Fprintf(w, string(b))
}
func PostTodoHtmlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, _ := template.ParseFiles("templates/posttodo.html", "templates/layout.html")
		tmpl.ExecuteTemplate(w, "layout", nil)
		w.WriteHeader(http.StatusOK)

	} else {
		//POST
		r.ParseForm()
		// fmt.Fprintln(w, r.Form)
		// fmt.Fprintln(w, "(1)", r.FormValue("todo_text"))
		// fmt.Fprintln(w, "(2)", r.PostFormValue("todo_text"))
		// fmt.Fprintln(w, "(3)", r.PostForm)

		ctx := context.Background()
		// conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
		conn, err := grpc.DialContext(ctx,
			TodoSvcAddr,
			grpc.WithInsecure(),
			// grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
		)
		if err != nil {
			panic(errors.Wrapf(err, "grpc: failed to connect %s", TodoSvcAddr))
		}

		defer conn.Close()
		client := pb.NewTodosClient(conn)
		text := &pb.Text{
			Text: r.PostFormValue("todo_text"),
		}
		todo, err := client.AddTodo(ctx, text)
		if err != nil {
			renderHTTPError(r, w, err, 500)
		}
		log.Printf("todo has been added: %+v", todo)
		http.Redirect(w, r, "/todos", 301)
	}

}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx,
		TodoSvcAddr,
		grpc.WithInsecure(),
		// grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
	)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", TodoSvcAddr))
	}

	defer conn.Close()
	client := pb.NewTodosClient(conn)
	t, err := client.ListTodos(ctx, &pb.Void{})
	if err != nil {
		renderHTTPError(r, w, err, 500)
	}
	tmpl, _ := template.ParseFiles("templates/todo.html", "templates/layout.html")
	tmpl.ExecuteTemplate(w, "layout", t)
	w.WriteHeader(http.StatusOK)
}

// func TodoHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := context.Background()
// 	// conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
// 	conn, err := grpc.DialContext(ctx,
// 		TodoSvcAddr,
// 		grpc.WithInsecure(),
// 		// grpc.WithStatsHandler(&ocgrpc.ClientHandler{})
// 	)
// 	if err != nil {
// 		panic(errors.Wrapf(err, "grpc: failed to connect %s", TodoSvcAddr))
// 	}

// 	defer conn.Close()
// 	client := pb.NewTodosClient(conn)
// 	t, err := client.ListTodos(ctx, &pb.Void{})
// 	if err != nil {
// 		renderHTTPError(r, w, err, 500)
// 	}
// 	for _, t := range t.Todos {
// 		fmt.Fprintf(w, "%d %s ", t.Id, t.Text)
// 		if t.Done {
// 			fmt.Fprintf(w, "✅")
// 		} else {
// 			fmt.Fprintf(w, "❌")
// 		}
// 		fmt.Fprintf(w, "\n")
// 	}
// 	w.WriteHeader(http.StatusOK)
// }
// func TodoCategoryHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "Todo[Category]: %v\n", vars["category"])
// }
// func TodoCategoriesHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "Todo: %v\n", vars)
// }
// func StorageHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "Storage: %v\n", vars["todo"])
// }
