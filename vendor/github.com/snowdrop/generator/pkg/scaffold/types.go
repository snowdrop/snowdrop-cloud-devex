package scaffold

type Project struct {
	GroupId            string
	ArtifactId         string
	Version            string
	PackageName        string
	Dependencies	   []string
	OutDir             string

	SnowdropBomVersion string
	SpringVersion      string
	Modules            []Module

	UrlService  	   string
}

type Config struct {
	Modules      []Module
}

type Module struct {
	Name	     string
	Description  string
	Starters     []Starter
}

type Starter struct {
	GroupId	     string
	ArtifactId	 string
	Scope	     string
}
