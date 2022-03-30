package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var paused = false
var curMode = 0
var sortField = 1
var reverse = true

type Result []string
type bySortfield []Result

func (r bySortfield) Len() int {
	return len(r)
}

func (r bySortfield) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r bySortfield) Less(i, j int) bool {
	va := r[i][sortField-1]
	vb := r[j][sortField-1]

	iA, erri_a := strconv.Atoi(va)
	iB, erri_b := strconv.Atoi(vb)
	fA, errf_a := strconv.ParseFloat(va, 64)
	fB, errf_b := strconv.ParseFloat(vb, 64)
	if errf_a == nil && errf_b == nil {
		if reverse {
			return !(fA < fB)
		} else {
			return fA < fB
		}
	} else if erri_a == nil && erri_b == nil {
		if reverse {
			return !(iA < iB)
		} else {
			return iA < iB
		}
	} else {
		if reverse {
			return !(va < vb)
		} else {
			return va < vb
		}
	}
}

func runCmd(program string, args []string) string {
	cmd := exec.Command(program, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		termbox.Close()
		fmt.Println(string(out))
		log.Printf("cmd.Run() failed with %s\n", err)
		os.Exit(1)
	}

	return string(out)
}

func parseResult(mode Mode, line string) Result {
	re := *regexp.MustCompile(mode.MatchRe)
	results := re.FindStringSubmatch(line) // , len(mode.Fields))
	values := results[1:]
	return Result(values)
}

func fmtValues(mode Mode, values []string) string {
	valstr := ""
	for i, v := range values {
		if mode.Fields[i].Hide {
			continue
		}

		width := 10
		if mode.Fields[i].Width != nil {
			width = *mode.Fields[i].Width
		}

		fmtStr := fmt.Sprintf("%%-%ds", width)
		valstr = valstr + " " + fmt.Sprintf(fmtStr, v)
	}
	return valstr
}

func fmtResult(mode Mode, result Result) string {
	valstr := ""
	for i, v := range result {
		if mode.Fields[i].Hide {
			continue
		}

		width := 10
		if mode.Fields[i].Width != nil {
			width = *mode.Fields[i].Width
		}

		if len(v) > width {
			v = v[0:width]
		}
		fmtStr := fmt.Sprintf("%%-%ds", width)
		valstr = valstr + " " + fmt.Sprintf(fmtStr, v)
	}
	return valstr
}

func drawModes(config Config) {
	offset := 0
	for i, m := range config.Modes {
		if i == curMode {
			tbprint(offset, 0, termbox.ColorBlack, termbox.ColorWhite, m.Name)
		} else {
			tbprint(offset, 0, termbox.ColorWhite, termbox.ColorBlack, m.Name)
		}
		offset += len(m.Name) + 1
	}
}

func redraw(mode Mode) {
	header := fmtValues(mode, mode.FieldNames())
	tbprint(0, 1, termbox.ColorBlack, termbox.ColorWhite, header)

	output := runCmd(mode.Cmd, mode.Args)
	lines := strings.Split(output, "\n")

	results := make([]Result, 0)
	for i, l := range lines {
		if len(l) == 0 {
			continue
		}

		if i < mode.DropHeader {
			continue
		}

		if i > len(lines)-mode.DropFooter-1 {
			break
		}

		result := parseResult(mode, strings.TrimSpace(l))
		results = append(results, result)
	}

	sort.Sort(bySortfield(results))

	for i, result := range results {
		valstr := fmtResult(mode, result)
		tbprint(0, i+2, termbox.ColorDefault, termbox.ColorDefault, valstr)
	}

	termbox.Flush()
}

func handleInput(config Config) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeySpace:
				paused = !paused
			default:
				if ev.Ch == 'q' {
					termbox.Close()
					os.Exit(0)
				}

				if ev.Ch >= '0' && ev.Ch <= '9' {
					newSortField, _ := strconv.Atoi(string(ev.Ch))
					if sortField == newSortField {
						reverse = !reverse
					}
					config.Modes[curMode].SortField = newSortField
					sortField = newSortField
				}

				if ev.Ch == ']' {
					reverse = true
					curMode++
					if curMode > len(config.Modes)-1 {
						curMode = 0
					}
					sortField = config.Modes[curMode].SortField
				}

				if ev.Ch == '[' {
					reverse = true
					curMode--
					if curMode < 0 {
						curMode = len(config.Modes) - 1
					}
					sortField = config.Modes[curMode].SortField
				}
			}
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}

func main() {
	cmd := filepath.Base(os.Args[0])

	systemconfig := "/etc/" + cmd + ".yml"

	usr, uerr := user.Current()
	if uerr != nil {
		log.Fatalf("Error loading current user; %s\n", uerr)
	}
	userconfig := usr.HomeDir + "/." + cmd + ".yml"

	configpaths := []string{userconfig, systemconfig}
	config := ReadConfig(configpaths)
	err := termbox.Init()
	if err != nil {
		log.Printf("Oopps", err)
		os.Exit(1)
	}

	defer termbox.Close()

	go handleInput(config)

	for {
		m := config.Modes[curMode]
		if !paused {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			drawModes(config)
			redraw(m)
		}
		interval := time.Duration(1)
		if m.Interval != nil {
			interval = *m.Interval
		}
		time.Sleep(interval * time.Second)
	}
}

// This function is often useful:
func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}
