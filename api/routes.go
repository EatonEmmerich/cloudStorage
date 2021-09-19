package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type middleware func(http.HandlerFunc)http.HandlerFunc

var _ middleware = POST
func POST(next http.HandlerFunc)http.HandlerFunc{
	return func (resp http.ResponseWriter, req *http.Request){
		if req.Method != "POST"{
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next(resp, req)
	}
}

var _ middleware = PATCH
func PATCH(next http.HandlerFunc)http.HandlerFunc{
	return func (resp http.ResponseWriter, req *http.Request){
		if req.Method != "PATCH"{
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next(resp, req)
	}
}

var _ middleware = GET
func GET(next http.HandlerFunc)http.HandlerFunc{
	return func (resp http.ResponseWriter, req *http.Request){
		if req.Method != "GET"{
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next(resp, req)
	}
}

func New(dbc *sql.DB) (http.Server){
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", POST(uploadDocument(dbc)))
	mux.HandleFunc("/update", PATCH(updateDocument(dbc)))
	mux.HandleFunc("/documents", GET(listDocuments(dbc)))
	mux.HandleFunc("/document", GET(getDocument(dbc)))

	httpServer := http.Server{
		Handler: mux,
		Addr: "localhost:8084",
		ErrorLog: log.Default(),
	}
	return httpServer
}

// Read file from multipart body and store as authenticated user's document
func uploadDocument(dbc *sql.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		mp, err := req.MultipartReader()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		var documentIDs []int
		var userID int64
		for {
			mp, err := mp.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			documentID, err := documents.Upload(req.Context(), dbc, userID, mp)
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				resp.Write([]byte(err.Error()))
				log.Default().Println(err)
				return
			}
			documentIDs = append(documentIDs, documentID)
		}

		encoder := json.NewEncoder(resp)
		err = encoder.Encode(struct {
			DocumentIDs []int
		}{
			DocumentIDs: documentIDs,
		})
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}
		resp.WriteHeader(http.StatusOK)
	}
}

var updateDocumentURLRegex = regexp.MustCompile(`\/update/(:id)`)

func updateDocument(dbc *sql.DB) http.HandlerFunc{
	return func(resp http.ResponseWriter, req *http.Request) {

		mp, err := req.MultipartReader()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		documentID, err := strconv.ParseInt(req.URL.Query().Get("document_id"), 10,64)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		part, err := mp.NextPart()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		var userID int64
		err = documents.Update(req.Context(), dbc, documentID, userID, part)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}
		resp.WriteHeader(http.StatusOK)
	}
}


func listDocuments(dbc *sql.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		var userID int64
		docs , err := documents.ListDocuments(req.Context(), dbc, userID)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		docsJson, err := json.Marshal(docs)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		resp.WriteHeader(http.StatusOK)
		resp.Write(docsJson)
	}
}


func getDocument(dbc *sql.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		documentID, err := strconv.ParseInt(req.FormValue("document_id"), 10, 64)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		_, reader, err := documents.OpenDocument(req.Context(), dbc, documentID)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}

		resp.Header().Set("Content-Type", "image/png")
		_, err = io.Copy(resp, reader)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}
	}
}