package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func connectToDatabase() error {
	// PostgreSQL database connection string
	// connStr := "postgres://postgres:vinayak@vinayak/postgres?sslmode=disable"

	connStr := "postgres://root:if4WnKsx6Y4Fx1SQOWccZLTAxjFHTbkT@dpg-cn7253djm4es73bpk5e0-a.singapore-postgres.render.com/demo_vdjv"

	// Open a connection to the database
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Check if the connection is successful by pinging the database
	if err = database.Ping(); err != nil {
		return err
	}

	// Set the global variable to the connected database instance
	db = database

	fmt.Println("Connected to the PostgreSQL database.")
	return nil
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := strings.Join(r.Form["title"], "")
	body := strings.Join(r.Form["body"], "")

	_, err := db.Exec("INSERT INTO tasks (title, body) VALUES ($1, $2)", title, body)
	if err != nil {
		fmt.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := getTasks()
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("task-list.html"))
	tmpl.Execute(w, tasks)
}

func getTasks() ([]Task, error) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Body); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}
	return tasks, nil
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		row := db.QueryRow("select id, title, body from tasks where id = $1", id)

		var t Task

		err := row.Scan(&t.ID, &t.Title, &t.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("edit-task.html")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		r.ParseForm()
		id := strings.Join(r.Form["id"], "")
		title := strings.Join(r.Form["title"], "")
		body := strings.Join(r.Form["body"], "")

		_, err := db.Exec("UPDATE tasks SET title=$1, body=$2 where id=$3", title, body, id)

		if err != nil {
			fmt.Print(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := strings.Join(r.Form["id"], "")

	// Execute the SQL DELETE statement to delete the task from the database
	_, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// Redirect back to the index path
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	// Connect to the PostgreSQL database
	err := connectToDatabase()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Define HTTP routes
	http.HandleFunc("/tasks", createTaskHandler)

	http.HandleFunc("/delete", deleteTaskHandler)

	http.HandleFunc("/update", updateTaskHandler)

	// Define HTTP routes
	http.HandleFunc("/", indexHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the HTTP server
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
