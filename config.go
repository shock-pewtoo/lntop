package main

import(
    "io/ioutil"
    "log"
    "time"
    "gopkg.in/yaml.v2"
    "os"
)

type Field struct {
	Name string `yaml:"name"`
	Hide bool   `yaml:"hide"`
    Width *int  `yaml:"width"`
}

type Config struct {
	Modes []Mode `yaml:"modes"`
}

// Modes
type Mode struct {
	Cmd    string   `yaml:"cmd"`
	Args   []string   `yaml:"args"`
	MatchRe  string   `yaml:"matchre"`
    Interval *time.Duration  `yaml:"interval"`
	Fields []Field `yaml:"fields"`
	Name   string   `yaml:"name"`
    DropHeader int `yaml:"dropheader"`
    DropFooter int `yaml:"dropfooter"`
    SortField int `yaml:"sortfield"`
}

func ReadConfig(paths []string) (Config) {
    found := false
    var actualpath string
    for _, path := range paths {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            continue
        } else {
            found = true
            actualpath = path
            break
        }
    }

    if !found {
        log.Fatalf("No configuration file found in %v", paths)
        os.Exit(1)
    }

    contents, err := ioutil.ReadFile(actualpath)
    if err != nil {
        log.Fatalf("Error reading config file: ", err)
    }

    var c Config
    err = yaml.Unmarshal(contents, &c)
    if err != nil {
        log.Fatalf("Error parsing config file: ", err)
    }

    for i, _ := range c.Modes {
        if c.Modes[i].SortField == 0 {
            c.Modes[i].SortField = 1
        }
    }
    return c
}

func (mode Mode) FieldNames() ([]string) {
    names := make([]string, 0)
    for _, f := range mode.Fields {
        names = append(names, f.Name)
    }
    return names
}
