package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type HistoryRecord struct {
	line    string
	cmd     string
	args    []string
	flags   []string
	cmdName string
}

func main() {

	historyFilePath, err := determinHistoryFilePath()
	if err != nil {
		log.Fatalf("Problem determin you history file path : %+v", err)
	}

	log.Printf("Found history path: %s\n", historyFilePath)

	historyFile, err := os.Open(historyFilePath)

	if err != nil {
		log.Fatalf("Error opening file: %+v", err)
	}

	defer historyFile.Close()

	scanner := bufio.NewScanner(historyFile)
	historyRecords := make([]HistoryRecord, 1000) // default to hist limit / 2 ?

	for scanner.Scan() {
		record, err := createHistoryRecordFromLine(scanner.Text())

		if err != nil {
			lineNumber := len(historyRecords)
			fmt.Errorf("line: %d, error: %+v", lineNumber, err)
			continue
		}

		historyRecords = append(historyRecords, record)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	cmdCounter := make(map[string]int)
	for _, record := range historyRecords {
		incrementCommandCounter(cmdCounter, record.cmdName)
	}

	for key, val := range cmdCounter {
		fmt.Printf("%s : %d\n", key, val)
	}
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

	defaultPath := ".zhistory" // zsh for now
	path := fmt.Sprintf("%s/%s", home, defaultPath)

	return path, nil
}
