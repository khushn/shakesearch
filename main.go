package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const FILE_NAME = "completeworks.txt"
const MAX_SEARCH_LIMIT = 100

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index

	Titles []string

	// A map of titles as key, and value is the index where the titular chapter begins
	TitlesMap map[string]int
	// Reverse index of the above
	TitlesMapRev map[int]string
	// Sorted titke index
	SortedTitleIndex []int

	// All the paragraph boundaries
	ParaBoundaries [][]int
}

type SearchResult struct {
	Title       string `json:"bookTitle"`
	IsBook      bool   `json:"IsBookSection`
	MatchedText string `json:"matchedText"`
}

func main() {
	searcher := Searcher{}
	err := searcher.ReadTitlesAndParaBreaks(FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}	

	err = searcher.Load(FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	err = searcher.BuildTitleIndex()
	if err != nil {
		log.Fatal(err)
	}

	// Just to debug, if everything is fine
	// later on, this can be moved in go test code
	testFindTitleForPos(&searcher)

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

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results, _ := searcher.Search(query[0])
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

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}

	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string) ([]*SearchResult, int) {
	// regex for case ignore search
	regex, _ := regexp.Compile("(?i)" + query + "(?-i)")
	// idxs := s.SuffixArray.Lookup([]byte(query), -1)
	idxs := s.SuffixArray.FindAllIndex(regex, MAX_SEARCH_LIMIT)
	firstInd := -1
	var results []*SearchResult
	for _, idx := range idxs {
		if firstInd == -1 {
			firstInd = idx[0]
		}
		// Add Title info
		searchRes := SearchResult{}
		title := s.findTitleForGivenindexPosition(idx[0])
		if len(title) > 0 {
			// results = append(results, "Book: " + title + "\n")
			searchRes.IsBook = true
			searchRes.Title = title
		}
		// results = append(results, s.CompleteWorks[idx[0]-250:idx[0]+250])
		searchRes.MatchedText = s.CompleteWorks[idx[0]-250 : idx[0]+250]
		results = append(results, &searchRes)
	}
	return results, firstInd
}

// To read all the titles, between 'conteny' and first title repeat
// Also using it catch all the para begins and ends
// To show better result snippets
func (s *Searcher) ReadTitlesAndParaBreaks(filename string) error {
	var err error
	var titles []string
	titlesMap := make(map[string]bool)	
	inTOC := false
	titlesReadingOver := false
	prevLineLength := 0
	indexFromBeg := 0
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	prevLineEmpty := false
	inSection := false
	sectionBegin := -1
	for err == nil {
		var line []byte
		indexFromBeg += prevLineLength
		line, _, err = r.ReadLine()
		prevLineLength = len(line)
		strline := strings.TrimSpace(string(line))
		if strline == "" {
			if inSection {
				paraBound := make([]int,2)
				paraBound[0] = sectionBegin
				paraBound[1] = indexFromBeg
				s.ParaBoundaries = append(s.ParaBoundaries, paraBound)
			}
			inSection = false
			prevLineEmpty = true
			continue
		} 

		if !titlesReadingOver {
			if inTOC {
				// Is a title till it repeats
				_, ok := titlesMap[strline]
				if ok {
					// titles over
					titlesReadingOver = true
					continue
				}
				titlesMap[strline] = true
				titles = append(titles, strline)
			}

			//fmt.Printf("err: %v, line: %v\n", err, strline)
			if strline == "Contents" {
				fmt.Printf("Begin reading Title metadata")
				inTOC = true
			}
		} else {
			// fmt.Printf("prevLineEmpty: %v\n", prevLineEmpty)
			if prevLineEmpty {				
				inSection = true
				sectionBegin = indexFromBeg
			}
		}
		prevLineEmpty = false

	}
	s.Titles = titles

	fmt.Printf("no. of titles, %v, titles: %+v\n", len(s.Titles), s.Titles)
	fmt.Printf("no. of s.ParaBoundaries, %v, s.ParaBoundaries: %+v\n", len(s.ParaBoundaries), s.ParaBoundaries[0])
	return nil
}

// This takes in the titles and uses the already built index to
// Have a collection of
// Should be called after calling Load()
func (s *Searcher) BuildTitleIndex() error {
	if s.SuffixArray == nil {
		err := errors.New("Call Load() before calling BuildTitleIndex()")
		return err
	}

	s.TitlesMap = make(map[string]int)
	s.TitlesMapRev = make(map[int]string)

	for _, title := range s.Titles {
		idxs := s.SuffixArray.Lookup([]byte(title), 2)
		fmt.Printf("Debug title: %v, idxs: %v\n", title, idxs)
		// we are interested in the 2nd one
		if len(idxs) > 1 {
			ind := int(math.Max(float64(idxs[0]), float64(idxs[1])))
			s.TitlesMap[title] = ind
			s.TitlesMapRev[ind] = title
			// Note: No need to sort, as it comes already sorted, because initial Titles are kept in array
			s.SortedTitleIndex = append(s.SortedTitleIndex, ind)
		}
	}

	fmt.Printf("Debug s.TitlesMap: %+v\n", s.TitlesMap)
	fmt.Printf("Debug s.TitlesMapRev: %+v\n", s.TitlesMapRev)
	fmt.Printf("Debug s.SortedTitleIndex: %+v\n", s.SortedTitleIndex)
	return nil
}

// Find which title the search query pertains to
// All list of titles are in 10s. No harm in using Log N solution, as it may be invoked multiple times
func (s *Searcher) findTitleForGivenindexPosition(pos int) string {
	title := ""
	N := len(s.SortedTitleIndex)
	beg := 0
	end := N - 1
	i := (beg + end) / 2
	for beg <= end && i < N && i >= 0 {
		if s.SortedTitleIndex[i] <= pos && (i+1 < N && s.SortedTitleIndex[i+1] >= pos) {
			// position found
			title = s.TitlesMapRev[s.SortedTitleIndex[i]]
			break
		}
		if s.SortedTitleIndex[i] < pos {
			beg = i + 1
		} else {
			end = i - 1
		}
		i = (beg + end) / 2
		//fmt.Printf("beg: %v, end: %v, i: %v\n", beg, end, i)
	}

	fmt.Printf("Debug title index: %v, title: %v\n", i, title)
	return title
}

func testFindTitleForPos(searcher *Searcher) {
	// debug test 1
	pos := 2921 // THE SONNETS
	title := searcher.findTitleForGivenindexPosition(pos)
	fmt.Printf("title for index position: %v, is %v\n", pos, title)
	// debug test 2
	pos = 4740620 + 50 // THE TRAGEDY OF TITUS ANDRONICUS
	title = searcher.findTitleForGivenindexPosition(pos)
	fmt.Printf("title for index position: %v, is %v\n", pos, title)
	// debug test 3
	pos = 0 // before any title
	title = searcher.findTitleForGivenindexPosition(pos)
	fmt.Printf("title for index position: %v, is %v\n", pos, title)
	// debug test 4
	// give a huge index
	pos = 1e9
	title = searcher.findTitleForGivenindexPosition(pos)
	fmt.Printf("title for index position: %v, is %v\n", pos, title)
}
