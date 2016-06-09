package main

import (
	"log"
	"net/http"

	"github.com/satori/go.uuid"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/storage/v1"
)

type Config struct {
	Addr         string
	GoogleConfig string
	BucketName   string
	MaxLength    int64
	CacheControl string
}

var googleConf *jwt.Config

func main() {
	var err error

	if googleConf, err = google.JWTConfigFromJSON([]byte(config.GoogleConfig), storage.DevstorageReadWriteScope); err != nil {
		log.Fatal(err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"POST"},
		OptionsPassthrough: false,
		Debug:              false,
	})

	r := mux.NewRouter()
	r.HandleFunc("/{bucket}", uploadHandler).Methods("POST")

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(negroni.NewRecovery())
	n.Use(c)
	n.UseHandler(r)
	n.Run(config.Addr)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	bucket := v["bucket"]
	if bucket == "" {
		w.Write([]byte("invalid request"))
		return
	}

	if r.ContentLength == 0 {
		w.Write([]byte("invalid request"))
		return
	}

	if config.MaxLength > 0 && r.ContentLength > config.MaxLength {
		w.Write([]byte("filesize limited to 10MiB"))
		return
	}

	contentType := r.Header.Get("Content-Type")
	fileName := uuid.NewV4().String()

	fs := r.Body
	storageService, err := storage.New(googleConf.Client(context.Background()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	object := &storage.Object{
		Name:         fileName,
		ContentType:  contentType,
		CacheControl: config.CacheControl,
	}
	res, err := storageService.Objects.Insert(bucket, object).Media(fs).Do()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("https://storage.googleapis.com/" + bucket + "/" + fileName))

	log.Printf("[uploaded]: %s\n", res.MediaLink)
}
