package main

import (
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/snowdrop/k8s-supervisor/pkg/scaffold"
	log "github.com/sirupsen/logrus"

	"io/ioutil"
	"io"
	"fmt"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/template/{id}", GetProject).Methods("GET")
	//router.HandleFunc("/zip", GetZip).Methods("GET")
	router.HandleFunc("/gzip", GetGZip).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	p := scaffold.Project{
		GroupId: "me.snowdrop",
		ArtifactId: "cool",
		Version: "1.0",
		PackageName: "io.openshift",
		SnowdropBomVersion: "1.5.15.Final",
		SpringVersion: "1.5.15.Release",
	}
	log.Infof("Params : ",params)
	json.NewEncoder(w).Encode(p)
}

func GetGZip(w http.ResponseWriter, r *http.Request) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	// Files to Gzip
	files := []string{"data/example.csv", "data/data.csv"}
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Add files to zip
	for _, file := range files {
		b, _ := ioutil.ReadFile(file)
		gz.Write(b)
	}
}

func GetZip(w http.ResponseWriter, r *http.Request) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	// Files to Zip
	files := []string{"data/example.csv", "data/data.csv"}

	errZip := zipFiles(w,files)
	if errZip != nil {
		log.Fatal(errZip)
	}
}

func zipFiles(w http.ResponseWriter, files []string) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}