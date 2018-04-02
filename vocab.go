package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

type sample struct {
	score int
	query string
}

func loadScores(scorePath string) (map[string]int, error) {
	if _, err := os.Stat(scorePath); os.IsNotExist(err) {
		return make(map[string]int), nil
	}
	bytes, err := ioutil.ReadFile(scorePath)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int)
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func saveScores(scorePath string, score map[string]int) error {
	outFile, err := os.Create("scores.json")
	if err != nil {
		return err
	}
	defer outFile.Close()
	jsonEnc := json.NewEncoder(outFile)
	return jsonEnc.Encode(score)
}

func getAns(stdin *bufio.Reader) (bool, error) {
	for true {
		s, err := stdin.ReadString('\n')
		s = strings.Trim(s, "\n")
		if err != nil {
			return false, err
		}
		if s == "y" || s == "yes" {
			return true, nil
		} else if s == "n" || s == "no" {
			return false, nil
		}
	}
	panic("never happens")
}

func readCSV(filename string) [][]string {
	file, err := os.Open(filename)
	content := make([]byte, 100)
	s := ""
	for true {
		length, err := file.Read(content)
		if err != nil {
			break
		}
		s += string(content[:length])
	}
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = ':'
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return records
}

func main() {
	records := readCSV("vocab.csv")
	inv := flag.Bool("inv", false, "Inverse mode")
	flag.Parse()

	queriesAnswers := make(map[string]map[string]bool)
	answers := make(map[string][]string)
	queries := make([]string, 0)
	for _, row := range records {
		var query, answer string
		if *inv == true {
			query, answer = row[1], row[0]
		} else {
			query, answer = row[0], row[1]
		}
		fmt.Printf("query: %s, answer: %s\n", query, answer)
		if queriesAnswers[query] == nil {
			queriesAnswers[query] = make(map[string]bool)
			queries = append(queries, query)
		}
		queriesAnswers[query][answer] = true
		answers[query] = append(answers[query], answer)
	}
	stdin := bufio.NewReader(os.Stdin)

	scores, _ := loadScores("scores.json")

	samples := make([]sample, len(queries))
	for i, q := range queries {
		samples[i] = sample{scores[q], q}
	}
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].score < samples[j].score
	})

	// create a channel waiting for the SIGTERM (ctrl + c)
	// catch it, save scores to json and quit nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		saveScores("scores.json", scores)
		os.Exit(0)
	}()

	for _, i := range rand.Perm(len(queries)) {
		query := queries[i]
		fmt.Printf("%s (%d): ", query, len(queriesAnswers[query]))
		stdin.ReadString('\n')
		ls := answers[query]
		for _, word := range ls[:len(ls)-1] {
			fmt.Printf("%s | ", word)
		}
		fmt.Println(ls[len(ls)-1])
		ans, err := getAns(stdin)
		if err != nil {
			panic("error")
		}
		// for corrent answer increment the score
		if ans {
			scores[query]++
		}
	}
	saveScores("scores.json", scores)
}
