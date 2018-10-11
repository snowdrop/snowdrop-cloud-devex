package cmd

import (
	"archive/zip"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/scaffold"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	c = &scaffold.Config{}
	p = scaffold.Project{
		GroupId:    "com.example",
		ArtifactId: "demo",
		Version:    "0.0.1-SNAPSHOT",
	}
)

const (
	ServiceEndpoint = "http://spring-boot-generator.195.201.87.126.nip.io"
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

			form := url.Values{}

			prompt := promptui.Select{
				Label: "Which type of project do you want to create?",
				Items: []string{"Template", "Custom"},
			}

			i, _, err := prompt.Run()
			if i == 0 {
				templates := &promptui.SelectTemplates{
					Active:   "\U0001F3ED {{ .Name | cyan }}",
					Inactive: "  {{ .Name | cyan }}",
					Selected: "\U0001F3ED {{ .Name | red | cyan }}",
					Details: `
--------- Template ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}`,
				}

				uiTemplates := getUiTemplates()
				prompt = promptui.Select{
					Label:     "Which template should we use?",
					Items:     uiTemplates,
					Templates: templates,
				}

				i, _, _ := prompt.Run()

				form.Add("template", uiTemplates[i].Name)

			} else {
				var lastSelected string
				var selectedModules []string
				for {
					templates := &promptui.SelectTemplates{
						Active:   "\U0001F680 {{ .Name | cyan }}",
						Inactive: "  {{ .Name | cyan }}",
						Selected: "\U0001F680 {{ .Name | red | cyan }}",
						Help:     fmt.Sprintf("Selected modules: %s. Select 'Done' when done.", selectedModules),
						Details: `
--------- Module ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}`,
					}
					modules := getUiModules(selectedModules)
					prompt = promptui.Select{
						Label:     "Which module(s) should we use?",
						Items:     modules,
						Templates: templates,
					}
					i, _, _ := prompt.Run()

					lastSelected = modules[i].Name

					if lastSelected == "Done" {
						break
					} else {
						selectedModules = append(selectedModules, lastSelected)
						form.Add("module", lastSelected)
					}
				}
			}

			artifactName := p.ArtifactId

			prompt = promptui.Select{
				Label: "Do you want to further customize your project?",
				Items: []string{"Yes", "No"},
			}
			_, answer, _ := prompt.Run()
			if answer == "Yes" {
				templates := &promptui.SelectTemplates{
					Active:   "\U0001F343 {{ .SpringBootVersion | cyan }}",
					Inactive: "  {{ .SpringBootVersion | cyan }}",
					Selected: "\U0001F343 {{ .SpringBootVersion | red | cyan }}",
					Details: `
---------  Spring Boot Version ----------
{{ "Name:" | faint }}	{{ .SpringBootVersion }}
{{ "Community:" | faint }}	{{ .Community }}
{{ "Snowdrop:" | faint }}	{{ .Snowdrop }}`,
				}

				prompt = promptui.Select{
					Label:     "Which Spring Boot version should we use?",
					Items:     getUiBoms(),
					Templates: templates,
				}

				result, _, err := prompt.Run()
				if err != nil {
					panic(err)
				}
				sb := c.Boms[result]

				boms := []string{sb.Community, sb.Snowdrop}
				prompt = promptui.Select{
					Label: "What flavor do you want to use?",
					Items: boms,
				}

				_, bom, _ := prompt.Run()
				form.Add("snowdropbom", bom)

				var entry promptui.Prompt
				var chosen string
				var packageName string

				entry = promptui.Prompt{
					Label:     "Group Id",
					Default:   p.GroupId,
					AllowEdit: true,
				}
				chosen, _ = entry.Run()
				packageName += chosen
				form.Add("groupid", chosen)

				entry = promptui.Prompt{
					Label:     "Artifact Id",
					Default:   artifactName,
					AllowEdit: true,
				}
				artifactName, _ = entry.Run()
				packageName += "." + artifactName
				form.Add("artifactid", artifactName)

				entry = promptui.Prompt{
					Label:     "Version",
					Default:   p.Version,
					AllowEdit: true,
				}
				chosen, _ = entry.Run()
				form.Add("version", chosen)

				entry = promptui.Prompt{
					Label:     "Package name",
					Default:   packageName,
					AllowEdit: true,
				}
				chosen, _ = entry.Run()
				form.Add("packagename", chosen)
			}

			currentDir, _ := os.Getwd()
			entry := promptui.Prompt{
				Label:     "Where should the project be created from the current directory?",
				Default:   artifactName,
				AllowEdit: true,
			}
			outDir, _ := entry.Run()

			log.Infof("Create command called")

			client := http.Client{}

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

			dir := filepath.Join(currentDir, outDir)
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

			/*
							apiVersion: component.k8s.io/v1alpha1
				kind: Component
				metadata:
				  name: my-spring-boot
				spec:
				  deployment: innerloop
				  runtime: springboot
				  version: 1.5.16
				  envs:
				  - name: SPRING_PROFILES_ACTIVE
				    value: openshift-catalog
			*/

			component := Component{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Component",
					APIVersion: "component.k8s.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-spring-boot",
				},
				Spec: ComponentSpec{
					ExposeService:  true,
					DeploymentMode: "innerloop",
					Runtime:        "springboot",
					Version:        "1.5.16",
					Envs: []Env{
						{
							Name:  "SPRING_PROFILES_ACTIVE",
							Value: "openshift-catalog",
						},
					},
				},
			}
			b, err := yaml.Marshal(component)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile("component.yml", b, 0644)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	createCmd.Flags().StringVarP(&p.UrlService, "urlservice", "u", ServiceEndpoint, "URL of the HTTP Server exposing the spring boot service")

	// Add a defined annotation in order to appear in the help menu
	createCmd.Annotations = map[string]string{"command": "create"}

	rootCmd.AddCommand(createCmd)
}

func getUiTemplates() []scaffold.Template {
	return templatesExceptNamed("custom")
}

func getUiModules(selectedModules []string) []scaffold.Module {
	modules := modulesExceptNamed(selectedModules...)
	return append(modules, scaffold.Module{Name: "Done", Description: "Select when done"})
}

type uiBom struct {
	scaffold.Bom
	SpringBootVersion string
}

func getUiBoms() (boms []uiBom) {
	for _, bom := range c.Boms {
		boms = append(boms, uiBom{
			bom,
			bom.GetSpringBootVersion(),
		})
	}

	return boms
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

func templatesExceptNamed(names ...string) (templates []scaffold.Template) {
	excluded := make(map[string]bool)

	for _, name := range names {
		if len(name) > 0 {
			for _, template := range c.Templates {
				if name == template.Name {
					excluded[name] = true
				}
			}
		}
	}

	for _, template := range c.Templates {
		if !excluded[template.Name] {
			templates = append(templates, template)
		}
	}

	return templates
}

func GetGeneratorServiceConfig() {
	// Call the /config endpoint to get the configuration
	URL := strings.Join([]string{ServiceEndpoint, "config"}, "/")
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
