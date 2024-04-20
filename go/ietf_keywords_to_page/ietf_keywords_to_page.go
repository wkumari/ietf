// ietf_keywords_to_page
//
// This program reads a csv file of WG names and keywords, and created a page of the form
// Keyword -> WGs.

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// This is used for naming the config file, etc.
const programName = "ietf_keywords_to_page"

type options struct {
	infile   string
	overview bool

	debug   bool
	verbose bool
}

var (
	opts = options{}
)

func usage() {
	const msg = `
  Reads a CSV file of WGs and keywords, and outputs a webpage of keyword -> WG.

  It reads a YAML config file called 'config' in the current directory or
  in ~/.%v/

  Example config:

Flags:
`

	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Printf(msg, programName)
	flag.PrintDefaults()
}

func parseFlags() {
	flag.Usage = usage

	viper.SetConfigName("config")                // name of config file (without extension)
	viper.SetConfigType("yaml")                  // config file type. Is affed to config file name
	viper.AddConfigPath("$HOME/." + programName) // call multiple times to add many search paths
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Debug("No config file (config.yaml) found.")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	log.SetLevel(log.WarnLevel)

	// Flags:  long name, short name, default value, description
	flag.StringP("infile", "i", "", "input cvs file")
	flag.BoolP("overview", "o", false, "Generate overview page only (keywords containing *)")

	flag.BoolP("verbose", "v", false, "be more verbose.")
	flag.BoolP("debug", "d", false, "print debug information.")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	opts.infile = viper.GetString("infile")
	opts.overview = viper.GetBool("overview")

	opts.debug = viper.GetBool("debug")
	opts.verbose = viper.GetBool("verbose")

	if opts.debug {
		log.SetLevel(log.DebugLevel)
	} else if opts.verbose {
		log.SetLevel(log.InfoLevel)
	}

	if opts.infile == "" {
		log.Error("--infile is a required parameter\n\n")
		usage()
		os.Exit(1)
	}
}

// ReadCSV read the provided CSV file and "pivots" it.
// It expects a header line listing WGs, and then for each subsequent line
// keywords in the appropriate column.
// A,B,C
// 1,2,3
// ,1,1    ->
//
func ReadCSV(reader *bufio.Reader) map[string][]string {
	kw := make(map[string][]string)

	r := csv.NewReader(reader)
	r.FieldsPerRecord = 0
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse", err)
	}
	for row := 1; row < len(records); row++ {
		for i := range records[row] {
			keyword := records[row][i]
			// If there is a keyword (many columns are empty)
			if keyword != "" {
				if opts.overview == true {
					if strings.Contains(keyword, "*") == true {
						keyword = strings.Trim(keyword, "*")
						keyword = strings.TrimSpace(keyword)

						// If the first character is lowercase, assume not an acronym and TitleCase it.
						if string(keyword[0]) != strings.ToUpper(string(keyword[0])) {
							keyword = strings.Title(strings.ToLower(keyword))
						}
						log.Infof("%s is an overview keyword", keyword)
						kw[keyword] = append(kw[keyword], records[0][i])
					}
				} else {
					keyword = strings.Trim(keyword, "*")
					keyword = strings.TrimSpace(keyword)

					// If the first character is lowercase, assume not an acronym and TitleCase it.
					if string(keyword[0]) != strings.ToUpper(string(keyword[0])) {
						keyword = strings.Title(strings.ToLower(keyword))
					}
					kw[keyword] = append(kw[keyword], records[0][i])
				}
			}
		}
	}
	return kw
}

// GenerateHTML takes the keyword list and generates HTML.
func GenerateHTML(keywords map[string][]string) string {
	// The {{if $index}},{{end}} bit gets a comma between WGs.
	var header string

	if opts.overview == true {
		header = `
		<!DOCTYPE html>
	<html>
	  <head>
		<title>Keywords</title>
		<link rel="stylesheet" href="wagtail.css" type="text/css">
	  </head>
	  <body style="background-color:#F9FBFC;">
	  	<div class="body body-panel clearfix" style="margin-left: 150px;">
	  	<h1>Overview Keywords</h1>
		<p>This page helps you find which working groups are working on specific (high level) technologies.</br>
		It is designed to be a simple way for both newcomers to figure out where work they are interested in may
		be being discussed. This is a very high level overview, more detailed information is at <a href="page.html">over here</a>.<p>

		`
	} else {
		header = `
		<!DOCTYPE html>
		<html>
		  <head>
			<title>Keywords</title>
			<link rel="stylesheet" href="wagtail.css" type="text/css">
		  </head>
		  <body style="background-color:#F9FBFC;">
			  <div class="body body-panel clearfix" style="margin-left: 150px;">
			  <h1>Detail Keywords</h1>
			<p>This page has a simple mapping from various keywords and acronyms to IETF WGs that are working on these.</br>
			It is designed so established IETF particpants to figure out where
			new work should go, where a specific protocol or technology of interest is being discussed, etc.</br>
			A higher level / introductory page is <a href="index.html">over here</a>.<p>

		`
	}

	body := `

	<p><b>Note:</b> This is very much still a proof of concept / work in progress, and is *very* far from complete.
	</br></br>
	</p>
	<div class="rich-text">
		 <ul>
		  {{ with . }}
			{{ range .}}
				  <li>
					  <b>{{ .Keyword }}</b> -
					  {{ range $index, $group := .WGs}}{{if $index}},{{end}} <a href="https://datatracker.ietf.org/wg/{{ $group }}/about/">{{ $group }}</a>{{ end }}
				  </li>
			{{ end }}
		  {{ end }}
		  </ul>
		</div>
	</div>
		  <hr>
		  <small>
		  Warren Kumari (warren@kumari.net) is helping to organize these - if you know of a good keyword which maps to a WG, please let him know.</small>
  </body>
</html>
	`
	tmpl := header + body
	type kwmap struct {
		Keyword string
		WGs     []string
	}
	var result []kwmap
	var entry kwmap
	var wglist []string
	var keys []string

	// No sorted maps in Go, so manually sort
	for kw := range keywords {
		keys = append(keys, kw)
	}
	sort.Strings(keys)

	for _, kw := range keys {
		wglist = nil
		for wg := range keywords[kw] {
			wglist = append(wglist, keywords[kw][wg])
		}
		entry.Keyword = kw
		entry.WGs = wglist
		result = append(result, entry)
	}
	t, err := template.New("keywords").Parse(tmpl)
	err = t.Execute(os.Stdout, result)
	if err != nil {
		panic(err)
	}
	return "OK"
}

func main() {

	parseFlags()
	file, err := os.Open(opts.infile)
	if err != nil {
		log.Fatal(err, "unable to open %s", opts.infile)
	}
	defer file.Close()

	keywords := ReadCSV(bufio.NewReader(file))
	GenerateHTML(keywords)
}
