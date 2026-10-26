package mapping

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const UnknownCommand = "ERROR: unknown command"

type Table struct {
	items map[string]string
}

func Load(path string) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open mapping file %s: %w", path, err)
	}
	defer f.Close()

	items := make(map[string]string)
	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		req, resp, ok := strings.Cut(line, "\t")
		if !ok || req == "" {
			log.Printf("mapping %s:%d: skip invalid line", path, lineNum)
			continue
		}
		items[req] = resp
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read mapping file %s: %w", path, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("mapping file %s: no valid mappings", path)
	}
	return &Table{items: items}, nil
}

func (t *Table) Lookup(request string) string {
	if resp, ok := t.items[request]; ok {
		return resp
	}
	return UnknownCommand
}
