package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents"
	"io"
	"log"
	"net/http"
)

func New(dbc *sql.DB) (http.Server){
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadDocument(dbc))
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
				resp.WriteHeader(http.StatusBadRequest)
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
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			log.Default().Println(err)
			return
		}
		resp.WriteHeader(http.StatusOK)
	}
}
