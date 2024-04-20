#!/bin/bash

mv ~/Downloads/WG\ Keywords\ -\ Keyword\ with\ asterisk\ are\ intro\ _\ generic\ \ -\ Sheet1.csv keywords.csv
go run ietf_keywords_to_page.go --infile ./keywords.csv > page.html
go run ietf_keywords_to_page.go -o --infile ./keywords.csv > index.html
rm keywords.csv
scp *.html sim.kumari.net:~/tmp
