package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"io/ioutil"
	"net/http"
	"strings"
	"os"
	"path/filepath"
	"archive/zip"
	"io"
	"fmt"
)

var (
	template   string
	templates  = []string{"simple"}
)

var createCmd = &cobra.Command{
	Use:     "create [flags]",
	Short:   "Create a Spring Boot maven project",
	Long:    `Create a Spring Boot maven project".`,
	Example: ` sb create`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var valid bool
		/*
		p := scaffold.Project{
			GroupId: "me.snowdrop",
			ArtifactId: "cool",
			Version: "1.0",
			PackageName: "io.openshift",
			SnowdropBomVersion: "1.5.15.Final",
			SpringVersion: "1.5.15.Release",
		}
		*/

		for _, t := range templates {
			if template == t {
				valid = true
			}
		}

		if !valid {
			log.WithField("template", mode).Fatal("The provided template is not supported: ")
		}

		log.Info("Create command called with template '%s'", template)

		client := http.Client{}

		u := "http://localhost:8000/template/simple"
		req, err := http.NewRequest(http.MethodGet, u, strings.NewReader(""))

		if err != nil {
			log.Error(err.Error())
		}
		addClientHeader(req)

		res, err := client.Do(req)
		if err != nil {
			log.Error(err.Error())
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error(err.Error())
		}

		currentDir, _ := os.Getwd()
		dir := filepath.Join(currentDir, "generated")
		zipFile := dir + ".zip"

		err = ioutil.WriteFile(zipFile, body, 0644)
		if err != nil {
			log.Errorf("Failed to download file %s due to %s", zipFile, err)
		}
		err = Unzip(zipFile, dir)
		if err != nil {
			log.Errorf("Failed to unzip new project file %s due to %s", zipFile, err)
		}
		err = os.Remove(zipFile)
		if err != nil {
			log.Errorf(err.Error())
		}

	},
}

func init() {
	createCmd.Flags().StringVarP(&template, "template", "t", "",
		fmt.Sprintf("Template name used to select the project to be created. Supported templates are '%s'", strings.Join(templates, ",")))
	createCmd.MarkFlagRequired("template")
	// Add a defined annotation in order to appear in the help menu
	createCmd.Annotations = map[string]string{"command": "create"}

	rootCmd.AddCommand(createCmd)
}

func addClientHeader(req *http.Request) {
	// TODO Define a version
	userAgent := "sb/1.0"
	req.Header.Set("User-Agent", userAgent)
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		name := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(name, os.ModePerm)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(name, string(os.PathSeparator)); lastIndex > -1 {
				fdir = name[:lastIndex]
			}

			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(
				name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}