package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents"
	"github.com/EatonEmmerich/cloudStorage/pkg/users"
	"github.com/EatonEmmerich/cloudStorage/pkg/users/authentication"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

var _ middleware = post

func post(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(resp, req)
	}
}

var _ middleware = patch

func patch(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != "PATCH" {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(resp, req)
	}
}

var _ middleware = get

func get(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(resp, req)
	}
}

type authorisedHandler func(userID int64, resp http.ResponseWriter, req *http.Request)

func basicAuth(dbc *sql.DB, next authorisedHandler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			resp.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		}
		userID, err := authentication.BasicAuthentication(req.Context(), dbc, username, password)
		if err != nil {
			respondError(resp, err)
			return
		}
		next(userID, resp, req)
	}
}

func New(dbc *sql.DB) http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", post(registerUser(dbc)))
	mux.HandleFunc("/upload", post(basicAuth(dbc, uploadDocument(dbc))))
	mux.HandleFunc("/update", patch(basicAuth(dbc, updateDocument(dbc))))
	mux.HandleFunc("/documents", get(basicAuth(dbc, listDocuments(dbc))))
	mux.HandleFunc("/document", get(basicAuth(dbc, getDocument(dbc))))
	//mux.HandleFunc("/")

	httpServer := http.Server{
		Handler:  mux,
		Addr:     "localhost:8084",
		ErrorLog: log.Default(),
	}
	return httpServer
}

func registerUser(dbc *sql.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		userID, err := users.Register(req.Context(), dbc, username, password)
		if err != nil {
			respondError(resp, err)
			return
		}
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("userID: " + strconv.Itoa(int(userID))))
	}
}

func respondError(resp http.ResponseWriter, err error) {
	if errors.Is(err, access_control.ErrAccessDenied){
		resp.WriteHeader(http.StatusUnauthorized)
	} else {
		resp.WriteHeader(http.StatusBadRequest)
	}
	log.Default().Println(err)
	_, respErr := io.WriteString(resp, err.Error())
	if respErr != nil {
		log.Default().Println(respErr)
	}
}

// Read file from multipart body and store as authenticated user's document
func uploadDocument(dbc *sql.DB) authorisedHandler {
	return func(userID int64, resp http.ResponseWriter, req *http.Request) {
		mp, err := req.MultipartReader()
		if err != nil {
			respondError(resp, err)
			return
		}

		var documentIDs []int
		for {
			part, err := mp.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			documentID, err := documents.Upload(req.Context(), dbc, userID, part, part.Header.Get("Content-Type"), part.FileName())
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				io.WriteString(resp, err.Error())
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
			respondError(resp, err)
			return
		}
	}
}

var updateDocumentURLRegex = regexp.MustCompile(`\/update/(:id)`)

func updateDocument(dbc *sql.DB) authorisedHandler {
	return func(userID int64, resp http.ResponseWriter, req *http.Request) {

		mp, err := req.MultipartReader()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		documentID, err := strconv.ParseInt(req.URL.Query().Get("document_id"), 10, 64)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		part, err := mp.NextPart()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		err = documents.Update(req.Context(), dbc, documentID, userID, part, part.Header.Get("Content-Type"), part.FileName())
		if err != nil {
			respondError(resp, err)
			return
		}
	}
}

func listDocuments(dbc *sql.DB) authorisedHandler {
	return func(userID int64, resp http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		docs, err := documents.ListDocuments(req.Context(), dbc, userID)
		if err != nil {
			respondError(resp, err)
			return
		}

		docsJson, err := json.Marshal(docs)
		if err != nil {
			respondError(resp, err)
			return
		}

		resp.WriteHeader(http.StatusOK)
		resp.Write(docsJson)
	}
}

func getDocument(dbc *sql.DB) authorisedHandler {
	return func(userID int64, resp http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		documentID, err := strconv.ParseInt(req.FormValue("document_id"), 10, 64)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			log.Default().Println(err)
			_, respErr := io.WriteString(resp, err.Error())
			log.Default().Println(respErr)
			return
		}

		doc, reader, err := documents.OpenDocument(req.Context(), dbc, documentID, userID)
		if err != nil {
			respondError(resp, err)
			return
		}

		resp.Header().Set("Content-Disposition", "attachment; filename=\""+html.EscapeString(doc.FileName)+"\"")
		resp.Header().Set("Content-Type", doc.MediaType)
		_, err = io.Copy(resp, reader)
		if err != nil {
			respondError(resp, err)
			return
		}
	}
}
