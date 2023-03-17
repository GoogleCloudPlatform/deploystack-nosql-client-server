// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//go:embed static
var static embed.FS

const (
	mongoport   = "27017"
	defaultPort = "80"
)

func main() {
	dbHost := os.Getenv("DBHOST")
	port := os.Getenv("PORT")

	if port == "" {
		port = defaultPort
	}

	if dbHost == "" {
		log.Fatal("DBHOST env not set. Need ip/host with mongo on port", mongoport)
	}

	ctx := context.Background()
	svc, err := newTrainerService(ctx, dbHost, mongoport)
	if err != nil {
		log.Fatal("error connecting to mongo: ", err)
	}

	fSys, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/healthz", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/healthz", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/trainer", listHandler(svc)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/trainer", createHandler(svc)).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/trainer", deleteHandler(svc)).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/trainer", updateHandler(svc)).Methods(http.MethodPut)
	router.PathPrefix("/").Handler(http.FileServer(http.FS(fSys)))

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

func respond(w http.ResponseWriter, r *http.Request, status int, content []byte, err error) {

	w.WriteHeader(status)
	if content == nil && err != nil {
		log.Error("", "status", status, "error", err)
		content = []byte(err.Error())
		w.Write(content)
		return
	}
	log.Info(r.Method, "status", status)
	w.Write(content)
	return
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusInternalServerError, []byte("ok"), nil)
	return
}

func listHandler(svc trainerCRUDer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trainers, err := svc.list(context.Background())
		if err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		j, err := json.Marshal(trainers)
		if err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}
		respond(w, r, http.StatusOK, j, nil)
		return
	}
}

func createHandler(svc trainerCRUDer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := trainer{}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		if err := svc.create(context.Background(), t); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		j, err := json.Marshal(t)
		if err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}
		respond(w, r, http.StatusCreated, j, nil)
		return
	}
}

func deleteHandler(svc trainerCRUDer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := trainer{}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		if err := svc.delete(context.Background(), t); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		respond(w, r, http.StatusNoContent, nil, nil)
		return
	}
}

func updateHandler(svc trainerCRUDer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]trainer{}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		log.Warn("", "data", data)

		if err := svc.update(context.Background(), data["original"], data["replacement"]); err != nil {
			respond(w, r, http.StatusInternalServerError, nil, err)
			return
		}

		respond(w, r, http.StatusOK, nil, nil)
		return
	}
}
