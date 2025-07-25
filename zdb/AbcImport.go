// AbcImport
package zdb

import (
	"fmt"
	"os"
	"path"
	"strings"
	"github.com/py60800/zique/zixml"
)

type AbcImporter struct {
}

func AbcImport(abc string) string {
	fmt.Println("Abc Import:", abc)
	txt := strings.Split(abc, "\n")

	var base string
	buffer := make([]byte, 0)
	for i, line := range txt {
		if strings.HasPrefix(line, "X:") {
			txt[i] = "X:1"
		}
		if strings.HasPrefix(line, "T:") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "T:"))
			base = NiceName(title)
		}
		buffer = append(buffer, []byte(txt[i])...)
		buffer = append(buffer, byte('\n'))
	}
	abcFile := path.Join("tmp", base+".abc")
	f, _ := os.Create(abcFile)
	f.Write(buffer)
	f.Close()

	xmlFile := Abc2Xml(abcFile, "./tmp")
	// Check for duplicate
	index := zixml.ComputeIndexForFile(xmlFile)
	MuseEdit(xmlFile)
	duplicates := tuneDB.GetDuplicates(index)
	fmt.Printf("Index:%s %s (%v)\n", index, xmlFile, duplicates)
	if len(duplicates) > 0 {
		warning := "Potential duplicates:\n"
		for i, d := range duplicates {
			warning += fmt.Sprintf("%d - %s\n", i, d)
			if i > 10 {
				break
			}
		}
		return warning
	}
	return ""
}
