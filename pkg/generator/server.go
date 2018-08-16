package main

import (
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/k8s-supervisor/pkg/scaffold"
	"github.com/snowdrop/k8s-supervisor/pkg/common/logger"


	"net/url"
)

var (
	currentDir, _    = os.Getwd()
	port			 = "8000"
	pathGeneratorDir = ""
	tmpDirName       = "_temp"
	letterRunes      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Check env vars
	s := os.Getenv("SERVER_PORT")
	if s != "" {
		port = s
	}

	t := os.Getenv("GENERATOR_PATH")
	if t != "" {
	   pathGeneratorDir = t
	}

	log.Infof("Starting HTTP Server on port %s, exposing endpoint %s",port,"/template/{id}")

	router := mux.NewRouter()
	router.HandleFunc("/template/{id}", GetProject).Methods("GET")

	log.Fatal(http.ListenAndServe(":" + port, router))
}

func getUrlVal(r *http.Request, k string) string {
	return r.URL.Query().Get(k)
}

func getArrayVal(r *http.Request, k string, params map[string][]string) []string {
	return params[k]
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	ids := mux.Vars(r)
	params, _ := url.ParseQuery(r.URL.RawQuery)

	p := scaffold.Project{
		GroupId: getUrlVal(r,"groupId"),
		ArtifactId: getUrlVal(r,"artifactId"),
		Version: getUrlVal(r,"version"),
		PackageName: getUrlVal(r,"packageName"),
		Dependencies: getArrayVal(r,"dependencies",params),
		SnowdropBomVersion: getUrlVal(r,"bomVersion"),
		SpringVersion: getUrlVal(r,"springbootVersion"),
		OutDir: getUrlVal(r,"outDir"),
	}
	log.Info("Project : ",p)
	log.Info("Params : ",ids)

	// Parse Starters Config YAML file to load the starters associated to a module (web, ...)
	scaffold.ParseStartersConfigFile(pathGeneratorDir)

	// Collect the java projects's template (simple, rest, ...) defined within the package template
	scaffold.CollectVfsTemplates(ids["id"])

	// Generate a random temp directory where populated files will be saved
	tmpdir := strings.Join([]string{tmpDirName,randStringRunes(10)}, "/")
	log.Infof("Temp dir %s",tmpdir)

	// Parse the java project's templates using the config of the project
	scaffold.ParseTemplates(currentDir,tmpdir,p)
	log.Info("Project generated")

	zipDir := strings.Join([]string{tmpdir,ids["id"],"/"},"/")
	handleZip(w,zipDir)
	log.Info("Zip populated")

	// Remove temp dir where project has been generated
	removeTempDir(tmpdir)
}

func removeTempDir(tmpdir string) {
	err := os.RemoveAll(strings.Join([]string{currentDir,tmpdir},"/"))
	if err != nil {
		log.Error(err.Error())
	}
}

// Generate Zip file to be returned as HTTP Response
func handleZip(w http.ResponseWriter,tmpdir string) {
	zipFilename := "generated.zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))

	errZip := zipFiles(w, tmpdir)
	if errZip != nil {
		log.Fatal(errZip)
	}
}

// Get Files generated from templates under _temp directory and
// them recursively to the file to be zipped
func zipFiles(w http.ResponseWriter,tmpdir string) error {
	fullPathZipDir := strings.Join([]string{currentDir,tmpdir},"/")
	log.Info("Zip file path : ",fullPathZipDir)
	err := recursiveZip(w,fullPathZipDir)
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

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

