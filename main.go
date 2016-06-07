package main

import (
	"log"
	"mime"
	"net/http"

	"github.com/satori/go.uuid"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/storage/v1"
)

type Config struct {
	Addr         string
	DevMode      bool
	ProjectID    string
	GoogleConfig string
	BucketName   string
	MaxLength    int64
}

var googleConf *jwt.Config

func main() {
	var err error

	if googleConf, err = google.JWTConfigFromJSON([]byte(config.GoogleConfig), storage.DevstorageReadWriteScope); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", uploadHandler).Methods("POST")

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(negroni.NewRecovery())
	n.UseHandler(r)
	n.Run(config.Addr)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if config.MaxLength > 0 && r.ContentLength > config.MaxLength {
		w.Write([]byte("filesize limited to 10MiB"))
		return
	}

	ext := ""
	if exts, err := mime.ExtensionsByType(r.Header.Get("Content-Type")); err == nil {
		ext = exts[0]
	}
	fileName := uuid.NewV4().String() + ext

	fs := r.Body
	storageService, err := storage.New(googleConf.Client(context.Background()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	object := &storage.Object{Name: fileName}
	res, err := storageService.Objects.Insert(config.BucketName, object).Media(fs).Do()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Write([]byte(res.MediaLink))

	log.Printf("[uploaded]: %s\n", res.MediaLink)
}
