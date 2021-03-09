package main

import(
    "io/ioutil"
    "log"
    "time"
    "gopkg.in/yaml.v2"
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
}

func ReadConfig(configfile string) (Config) {
    contents, err := ioutil.ReadFile(configfile)
    if err != nil {
        log.Fatalf("Error reading config file: ", err)
    }

    var c Config
    err = yaml.Unmarshal(contents, &c)
    if err != nil {
        log.Fatalf("Error parsing config file: ", err)
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
