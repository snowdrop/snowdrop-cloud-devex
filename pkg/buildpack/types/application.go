package types

type Application struct {
	Name    string
	Replica int
	Cpu     string  `default:"100m"`
	Memory  string  `default:"250Mi"`
	Port    int32   `default:"8080"`
	Image   Image
}

type Image struct {
	Name string
	Repo string
}
