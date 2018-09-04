package scaffold

type Project struct {
	GroupId            string
	ArtifactId         string
	Version            string
	PackageName        string
	OutDir             string
	Template 		   string       `yaml:"template"  json:"template"`

	SnowdropBomVersion string
	SpringBootVersion  string
	Modules            []string

	UrlService  	   string
}
