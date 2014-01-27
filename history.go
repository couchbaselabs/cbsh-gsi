package main

import (
	"bufio"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/liner"
	"os"
)

func LoadHistory(c *api.Context) {
	historyFile := c.Cursh.HistoryFile()
	ReadHistoryFromFile(c.Liner, historyFile)
}

func UpdateHistory(c *api.Context, line string) {
	c.Liner.AppendHistory(line)
	WriteHistoryToFile(c.Liner, c.Cursh.HistoryFile())
}

func WriteHistoryToFile(liner *liner.State, path string) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	if _, err = liner.WriteHistory(writer); err == nil {
		writer.Flush()
	}
	return
}

func ReadHistoryFromFile(liner *liner.State, path string) error {
	if f, err := os.Open(path); err != nil {
		return err
	} else {
		defer f.Close()
		reader := bufio.NewReader(f)
		liner.ReadHistory(reader)
	}
	return nil
}
