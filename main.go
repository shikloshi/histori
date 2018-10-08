package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type HistoryRecord struct {
	line    string
	cmd     string
	args    []string
	flags   []string
	cmdName string
	envVars map[string]string
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// benchmark different methods
// create a cli wrapper around this and query the history
func main() {

	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatalf("Please provide command and command name to count, you can also use all.")
	}

	cmdName := args[1]

	historyRecords := buildHistoryModel() // Cache it somewhere

	cmdsCount := countCommands(historyRecords)
	if cmdName == "all" {
		pl := toSortedPairList(cmdsCount)
		printAllCmd(&pl, 30) // dont do hard coded
	}

	count := getCommandCount(cmdsCount, cmdName)

	log.Printf("Command: %s was executed %d times", cmdName, count)
}

func toSortedPairList(cmdsCount map[string]int) {
	pl := make(PairList, len(cmdsCount))
	i := 0
	for k, v := range cmdsCount {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
}

func printAllCmd(sortedCmdsCount *PairList, threshold int) {
	for _, p := range *sortedCmdsCount {
		if p.Value > thresold {
			log.Printf("Cmd name: %s ran %d times", p.Key, p.Value)
		}
	}
}

func buildHistoryModel() []HistoryRecord {

	historyFilePath, err := determineHistoryFilePath()
	if err != nil {
		log.Fatalf("Problem determin you history file path : %+v", err)
	}

	log.Printf("Found history path: %s\n", historyFilePath)

	historyFile, err := os.Open(historyFilePath)
	defer historyFile.Close()

	if err != nil {
		log.Fatalf("Error opening file: %+v", err)
	}

	return parseHistoryFile(historyFile)
}

func countCommands(historyRecords []HistoryRecord) map[string]int {
	cmdsCounter := make(map[string]int)
	for _, record := range historyRecords {
		incrementCommandCounter(cmdsCounter, record.cmdName)
	}
	return cmdsCounter
}

func getCommandCount(cmdsCounter map[string]int, commandName string) int {
	if count, ok := cmdsCounter[commandName]; ok {
		return count
	}
	return 0
}

func parseHistoryFile(historyFile *os.File) []HistoryRecord {
	historyRecords := make([]HistoryRecord, 1000) // default to hist limit?

	scanner := bufio.NewScanner(historyFile)

	for scanner.Scan() {
		record, err := createHistoryRecordFromLine(scanner.Text())

		if err != nil {
			lineNumber := len(historyRecords)
			fmt.Errorf("line: %d, error: %+v", lineNumber, err)
			continue
		}

		historyRecords = append(historyRecords, record)
	}

	// Find a better place for this
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return historyRecords
}

func incrementCommandCounter(cmdCounter map[string]int, cmdName string) {
	val, ok := cmdCounter[cmdName]
	if !ok {
		val = 0
	}
	val += 1
	cmdCounter[cmdName] = val
}

func createHistoryRecordFromLine(line string) (HistoryRecord, error) {

	sl := strings.Split(line, ";")
	if len(sl) < 2 {
		return HistoryRecord{},
			errors.New(fmt.Sprintf("Could not parse line: %s", line))
	}

	parsedCommand := strings.Split(sl[1], " ")

	if len(parsedCommand) < 1 {
		return HistoryRecord{},
			errors.New(fmt.Sprintf("There is no command in line: %s", line))
	}

	// For now discarding command that start with tnev var - we can make it more
	// generally and create a pattern for "correct" command
	if strings.Contains(sl[1], "=") {
		return HistoryRecord{},
			errors.New(fmt.Sprintf("First part is not a command. skipping line: %s", line))
	}

	return HistoryRecord{
		line:    line,
		cmd:     sl[1],
		cmdName: parsedCommand[0],
	}, nil
}

func determineHistoryFilePath() (string, error) {
	home := os.Getenv("HOME")

	if len(home) == 0 {
		return "", errors.New("Empty HOME env var, set it to point to your user home direcotry.")
	}

	defaultFileName := ".zhistory" // zsh for now
	path := fmt.Sprintf("%s/%s", home, defaultFileName)

	return path, nil
}
