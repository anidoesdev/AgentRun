package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)
type Job struct{
	ID int `json:"id"`
	Prompt string `json:"prompt"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"createdAt"` 
}

//store holds all the jobs 
type Store struct {
	mu sync.Mutex
	jobs map[int]Job
	nextID int
}

func NewStore() *Store{
	return &Store{jobs: make(map[int]Job), nextID: 1}
}

func (s *Store) Create(prompt string) Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := Job{
		ID: s.nextID,
		Prompt: prompt,
		Status: "queued",
		CreatedAt: time.Now().UTC(),
	}
	s.jobs[job.ID] = job
	s.nextID++
	return job
}

func (s *Store) Get(id int) (Job, bool){
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]

	return job, ok
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func healthz(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("ok\n"));
}

func createJob(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid JSON"})
			return 
		}
		if body.Prompt == ""{
			writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid JSON"})
			return
		}
		job := store.Create(body.Prompt)
		writeJSON(w, http.StatusCreated, job)
	}
	
}

func getJob(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err!=nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id must be a number"})
			return
		}
		job, ok := store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}


func main() {
	store := NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /jobs",createJob(store))
	mux.HandleFunc("GET /jobs/{id}", getJob(store))

	log.Println("agentRun listening on port: 8090")
	log.Fatal(http.ListenAndServe(":8090",mux))
}
