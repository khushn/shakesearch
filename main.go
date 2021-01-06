package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"regexp"
)

const FILE_NAME = "completeworks.txt"

func main() {
	searcher := Searcher{}
	titles, err := readTitles(FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("no. of titles, %v, titles: %+v\n", len(titles), titles)

	err = searcher.Load(FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

// To read all the titles, between 'conteny' and first title repeat
func readTitles(filename string) ([]string, error) {
	var err error
	var titles []string
	titlesMap := make(map[string]bool)
	inTOC := false
	f, err := os.Open(filename)
	if err != nil {
		return titles, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for err == nil {
		var line []byte
		line, _, err = r.ReadLine()		
		strline := strings.TrimSpace(string(line))
		if strline == "" {
			continue
		}

		if inTOC {
			// Is a title till it repeats
			_, ok := titlesMap[strline]
			if ok {
				// titles over
				break
			}
			titlesMap[strline] = true
			titles = append(titles, strline)
		}

		fmt.Printf("err: %v, line: %v\n", err, strline)
		if strline == "Contents" {
			fmt.Printf("Begin reading Title metadata")
			inTOC = true
		}

	}
	return titles, nil
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string) []string {
	// regex for case ignore search
	regex, _ := regexp.Compile("(?i)" + query + "(?-i)")
	// idxs := s.SuffixArray.Lookup([]byte(query), -1)
	idxs := s.SuffixArray.FindAllIndex(regex, -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx[0]-250:idx[0]+250])
	}
	return results
}
