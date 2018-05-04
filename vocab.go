package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

const scoreDecay = 0.1

type sample struct {
	score float64
	query string
}

func loadScores(scorePath string) (map[string]float64, error) {
	if _, err := os.Stat(scorePath); os.IsNotExist(err) {
		return make(map[string]float64), nil
	}
	bytes, err := ioutil.ReadFile(scorePath)
	if err != nil {
		return nil, err
	}
	res := make(map[string]float64)
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i] = res[i] + rand.Float64()*0.01
	}
	return res, nil
}

func saveScores(scorePath string, score map[string]float64) error {
	fmt.Println("\nsaving")
	outFile, err := os.Create("scores.json")
	if err != nil {
		return err
	}
	for i := range score {
		score[i] = math.Max(score[i], 0)
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
	it := 0
	for _, q := range queries {
		if scores[q] < 2 {
			samples[it] = sample{scores[q], q}
			it++
		} else {
			// TODO: defer it to the end, if all words were checked
			// fmt.Println(q)
			scores[q] -= scoreDecay
		}
		if scores[q] < 0 {
			scores[q] = 0
		}
	}
	samples = samples[:it]
	queries = nil

	// create a channel waiting for the SIGTERM (ctrl + c)
	// catch it, save scores to json and quit nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		saveScores("scores.json", scores)
		os.Exit(0)
	}()

	rand.Seed(int64(time.Now().Nanosecond()))

	// mode := "random"
	mode := "weakest first"
	var perm []int
	if mode == "random" {
		perm = rand.Perm(len(samples))
	} else if mode == "weakest first" {
		sort.Slice(samples, func(i, j int) bool {
			return samples[i].score < samples[j].score
		})
	} else {
		panic("unsupported mode")
	}

	fmt.Printf("%d words to learn\n", len(samples))

	var query string
	for i := 0; i < len(samples); i++ {
		if mode == "random" {
			query = samples[perm[i]].query
		} else if mode == "weakest first" {
			query = samples[i].query
		} else {
			panic("unsupported mode")
		}
		scores[query] -= scoreDecay
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
		} else {
            scores[query] -= 0.5
        }
		fmt.Println()
	}

	saveScores("scores.json", scores)
}
