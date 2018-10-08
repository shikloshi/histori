package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	//cobra "github.com/spf13/cobra"
)

type HistoryRecord struct {
	line    string
	cmd     string
	args    []string
	flags   []string
	cmdName string
	envVars map[string]string
}

// benchmark different methods
// create a cli wrapper around this and query the history
func main() {
	historyRecords := buildHistoryModelFromFile()
	fmt.Println(historyRecords)
}

func buildHistoryModelFromFile() []HistoryRecord {

	historyFilePath, err := determinHistoryFilePath()
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

func coundCommands(historyRecords []*HistoryRecord) {
	cmdsCounter := make(map[string]int)
	for _, record := range historyRecords {
		incrementCommandCounter(cmdsCounter, record.cmdName)
	}
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

func determinHistoryFilePath() (string, error) {
	home := os.Getenv("HOME")

	if len(home) == 0 {
		return "", errors.New("Empty HOME env var, set it to point to your user home direcotry.")
	}

	defaultFileName := ".zhistory" // zsh for now
	path := fmt.Sprintf("%s/%s", home, defaultFileName)

	return path, nil
}
