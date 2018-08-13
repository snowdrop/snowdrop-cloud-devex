package main

import (
	"compress/gzip"
	"net/http"
	"os"
	"io/ioutil"
	"fmt"
	"path/filepath"
	"archive/zip"
	"io"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/k8s-supervisor/pkg/scaffold"
	"github.com/snowdrop/k8s-supervisor/pkg/common/logger"
)

var (
	files           []string
	tmpdir          = "/_temp/"
	currentDir, _   = os.Getwd()
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	router := mux.NewRouter()
	router.HandleFunc("/template/{id}", GetProject).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getUrlVal(r *http.Request, k string) string {
	return r.URL.Query().Get(k)
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	p := scaffold.Project{
		GroupId: getUrlVal(r,"groupId"),
		ArtifactId: getUrlVal(r,"artifactId"),
		Version: getUrlVal(r,"version"),
		PackageName: getUrlVal(r,"packageName"),
		SnowdropBomVersion: getUrlVal(r,"bomVersion"),
		SpringVersion: getUrlVal(r,"springbootVersion"),
		OutDir: getUrlVal(r,"outDir"),
	}
	log.Info("Project : ",p)
	log.Info("Params : ",params)

	scaffold.CollectBoxTemplates(params["id"])
	scaffold.ParseTemplates(currentDir,tmpdir,p)
	log.Info("Project generated")

	handleZip(w)
	log.Info("Zip populated")
}

// Generate Zip file to be returned as HTTP Response
func handleZip(w http.ResponseWriter) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	errZip := zipFiles(w)
	if errZip != nil {
		log.Fatal(errZip)
	}
}

// Get Files generated from templates under _temp directory and
// them recursively to the file to be zipped
func zipFiles(w http.ResponseWriter) error {
	err := recursiveZip(w,currentDir + tmpdir)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func recursiveZip(w http.ResponseWriter, destinationPath string) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	err := filepath.Walk(destinationPath, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(destinationPath))
		relPath = strings.TrimPrefix(relPath,"/")
		log.Debugf("relPath calculated : ",relPath)

		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = zipWriter.Close()
	if err != nil {
		return err
	}
	return nil
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		// it's a directory then we skip
		log.Infof("%s is a dir, we skip it",fi.Name())
		return nil, err
	} else {
		// it's not a directory
		log.Infof("Read : %s",fi.Name())
		return ioutil.ReadAll(f)
	}
}

func handleGZip(w http.ResponseWriter) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	// Files to Gzip
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Add files to zip
	for _, file := range files {
		b, _ := readFile(file)
		gz.Write(b)
	}
}

func GetGZip(w http.ResponseWriter, r *http.Request) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	// Files to Gzip
	tmpDir := "_temp/"
	files := []string{
		tmpDir + "pom.xml",
		tmpDir + "simple/RestApplication.java",
		tmpDir + "simple/service/Greeting.java",
		tmpDir + "simple/service/GreetingEndpoint.java",
	}
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Add files to zip
	for _, file := range files {
		b, _ := ioutil.ReadFile(file)
		gz.Write(b)
	}
}
