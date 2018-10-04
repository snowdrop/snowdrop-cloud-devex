package cmd

import (
	"github.com/posener/complete"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/scaffold"
	"github.com/spf13/cobra"
	"sort"

	"archive/zip"
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	c = &scaffold.Config{}
	p = scaffold.Project{}
)

const (
	SERVICE_ENDPOINT = "http://spring-boot-generator.195.201.87.126.nip.io"
)

func init() {

	// Call the service at this address SERVICE_ENDPOINT
	// to get the configuration
	GetGeneratorServiceConfig()

	createCmd := &cobra.Command{
		Use:     "create [flags]",
		Short:   "Create a Spring Boot maven project",
		Long:    `Create a Spring Boot maven project.`,
		Example: ` sd create`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			var valid bool
			if p.Template != "" {
				for _, t := range c.Templates {
					if p.Template == t.Name {
						valid = true
					}
				}
			} else {
				p.Template = "simple"
				valid = true
			}

			if !valid {
				log.WithField("template", p.Template).Fatal("The provided template is not supported: ")
			}

			log.Infof("Create command called")

			client := http.Client{}

			form := url.Values{}
			form.Add("template", p.Template)
			form.Add("groupid", p.GroupId)
			form.Add("artifactid", p.ArtifactId)
			form.Add("version", p.Version)
			form.Add("packagename", p.PackageName)
			form.Add("snowdropbom", p.SnowdropBomVersion)
			form.Add("springbootversion", p.SpringBootVersion)
			form.Add("outdir", p.OutDir)
			for _, v := range p.Modules {
				if v != "" {
					form.Add("module", v)
				}
			}

			parameters := form.Encode()
			if parameters != "" {
				parameters = "?" + parameters
			}

			u := strings.Join([]string{p.UrlService, "app"}, "/") + parameters
			log.Infof("URL of the request calling the service is %s", u)
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
			dir := filepath.Join(currentDir, p.OutDir)
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

	createCmd.Flags().StringVarP(&p.Template, "template", "t", "",
		fmt.Sprintf("Template name used to select the project to be created. Supported templates are '%s'", getTemplatesFromList()))
	createCmd.Flags().StringVarP(&p.UrlService, "urlservice", "u", SERVICE_ENDPOINT, "URL of the HTTP Server exposing the spring boot service")
	createCmd.Flags().StringArrayVarP(&p.Modules, "module", "m", []string{}, "Spring Boot modules/starters")
	createCmd.Flags().StringVarP(&p.GroupId, "groupid", "g", "", "GroupId : com.example")
	createCmd.Flags().StringVarP(&p.ArtifactId, "artifactid", "i", "", "ArtifactId: demo")
	createCmd.Flags().StringVarP(&p.Version, "version", "v", "", "Version: 0.0.1-SNAPSHOT")
	createCmd.Flags().StringVarP(&p.PackageName, "packagename", "p", "", "Package Name: com.example.demo")
	createCmd.Flags().StringVarP(&p.SpringBootVersion, "springbootversion", "s", "", "Spring Boot Version")
	createCmd.Flags().StringVarP(&p.SnowdropBomVersion, "snowdropbom", "b", "", "Snowdrop Bom Version")

	// Add a defined annotation in order to appear in the help menu
	createCmd.Annotations = map[string]string{"command": "create"}

	rootCmd.AddCommand(createCmd)
	Suggesters[GetFlagSuggesterName(createCmd, "snowdropbom")] = bomSuggester{}
	Suggesters[GetFlagSuggesterName(createCmd, "module")] = moduleSuggester{}
	Suggesters[GetFlagSuggesterName(createCmd, "template")] = templateSuggester{}
}

type bomSuggester struct {
}

func (i bomSuggester) Predict(args complete.Args) []string {
	var suggestions []string
	for _, bom := range c.Boms {
		sbVersion := bom.Community[:strings.LastIndex(bom.Community, ".")]
		suggestions = append(suggestions, sbVersion+".Community")
		suggestions = append(suggestions, sbVersion+".Snowdrop")
	}

	return suggestions
}

type moduleSuggester struct {
}

func (i moduleSuggester) Predict(args complete.Args) []string {
	var suggestions []string
	if strings.ContainsRune(args.Last, ',') {
		names := strings.Split(args.Last, ",")
		for _, mod := range modulesExceptNamed(names...) {
			suggestions = append(suggestions, args.Last+mod.Name)
		}
	} else {
		for _, mod := range c.Modules {
			suggestions = append(suggestions, mod.Name)
		}
	}

	sort.Strings(suggestions)
	return suggestions
}

func modulesExceptNamed(names ...string) (modules []scaffold.Module) {
	excluded := make(map[string]bool)

	for _, name := range names {
		if len(name) > 0 {
			for _, mod := range c.Modules {
				if name == mod.Name {
					excluded[name] = true
				}
			}
		}
	}

	for _, mod := range c.Modules {
		if !excluded[mod.Name] {
			modules = append(modules, mod)
		}
	}

	return modules
}

type templateSuggester struct {
}

func (i templateSuggester) Predict(args complete.Args) []string {
	var suggestions []string
	for _, template := range c.Templates {
		suggestions = append(suggestions, template.Name)
	}

	return suggestions
}

func GetGeneratorServiceConfig() {
	// Call the /config endpoint to get the configuration
	URL := strings.Join([]string{SERVICE_ENDPOINT, "config"}, "/")
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, URL, strings.NewReader(""))
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
	if strings.Contains(string(body), "Application is not available") {
		log.Fatal("Generator service is not able to find the config !")
	}
	err = yaml.Unmarshal(body, &c)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func addClientHeader(req *http.Request) {
	// TODO Define a version
	userAgent := "sb/1.0"
	req.Header.Set("User-Agent", userAgent)
}

func getTemplatesFromList() string {
	var buffer bytes.Buffer
	for i, t := range c.Templates {
		buffer.WriteString(t.Name)
		if i < len(c.Templates)-1 {
			buffer.WriteString(" ")
		}
	}
	return buffer.String()
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
