package types

type Application struct {
	Name    string
	Version string
	Namespace string
	Replica int
	Cpu     string  `default:"100m"`
	Memory  string  `default:"250Mi"`
	Port    int32   `default:"8080"`
	Image   Image
	SupervisordName string
}

type Image struct {
	Name string
	AnnotationCmds bool
	Repo string
}

func NewApplication() Application {
	return Application{
		Version: "1.0",
		Cpu: "100m",
		Memory: "250Mi",
		Replica: 1,
		Port: 8080,
		SupervisordName: "copy-supervisord",
	}
}