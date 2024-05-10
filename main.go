package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Post struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

var (
	posts   = make(map[int]Post)
	nextID  = 1
	postsMu sync.Mutex
)

func main() {
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/post/", postHandler)

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetPosts(r, w)
	case "POST":
		handlePostPosts(r, w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/post/"):])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
	}
	switch r.Method {
	case "GET":
		handleGetPost(r, w, id)
	case "POST":
		handleDeletePost(r, w, id)
	case "PUT":
		handleUpdatePost(r, w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetPosts(r *http.Request, w http.ResponseWriter) {
	postsMu.Lock()         // Allow only one request to be fullfilled at a time by locking the server
	defer postsMu.Unlock() // Unlocks the server after the function finishes executing

	postsSlice := make([]Post, 0, len(posts)) // Initializes a slice of type Post, with initial length of 0 and max capacity = length of posts

	for _, post := range posts { // Appendes all the posts to the slice
		postsSlice = append(postsSlice, post)
	}

	w.Header().Set("Content-Type", "application/json") // Sets header
	json.NewEncoder(w).Encode(postsSlice)              // Encodes response(which is postsSlice) to json
}

func handlePostPosts(r *http.Request, w http.ResponseWriter) {
	var post Post

	body, err := io.ReadAll(r.Body) // Reads the body content
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &post); err != nil { // Decodes the json and stores in the body variable in Post struct
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	postsMu.Lock()
	defer postsMu.Unlock()

	post.ID = nextID
	nextID++
	posts[post.ID] = post

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func handleGetPost(r *http.Request, w http.ResponseWriter, id int) {
	postsMu.Lock()
	defer postsMu.Unlock()

	post, ok := posts[id]
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func handleDeletePost(r *http.Request, w http.ResponseWriter, id int) {
	postsMu.Lock()
	defer postsMu.Unlock()

	_, ok := posts[id]
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
	}

	delete(posts, id)
	w.WriteHeader(http.StatusOK)
}

func handleUpdatePost(r *http.Request, w http.ResponseWriter, id int) {
	postsMu.Lock()
	defer postsMu.Unlock()

	post, ok := posts[id] // Get the post
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
	}

	body, err := io.ReadAll(r.Body) // Get the body
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &post); err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	posts[id] = post

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}
