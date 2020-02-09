package cfgparse

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

var (
	sectionHeaderRegexp = regexp.MustCompile("\\[([^]]+)\\]")
	keyValueRegexp = regexp.MustCompile("([^:=\\s][^:=]*)\\s*(?P<vi>[:=])\\s*(.*)$")
)

type section struct {
	items map[string]string
}


type CfgParser struct {
	fileType string
	sections map[string]section
	delimeter string
	mutex sync.Mutex
}


func New() *CfgParser {
	cfg := CfgParser{}
	return &cfg
}


func (c* CfgParser) ReadFile(filePath string) error {
	if len(filePath) == 0 {
		err := errors.New("File Name cannot be Empty")
		return err
	}
	cfgFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	c.Parse(cfgFile)
	return nil
}

func (c* CfgParser) Parse(cfgFile *os.File) {
	reader := bufio.NewReader(cfgFile)
	var lineNo int
	var curSection section
	var err error

	for err != nil {
		buff, _, err := reader.ReadLine()
		if err != nil{
			break
		}
		if len(buff) == 0 {
			continue
		}
		line := strings.TrimFunc(string(buff), unicode.IsSpace)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if isSection(line) {
			section := sectionHeaderRegexp.FindStringSubmatch(line)[1]
			fmt.Println("It is a section")
			fmt.Println(section)
		} else if isKeyValue(line) {
			sectionValue := keyValueRegexp.FindStringSubmatch(line)[1]
			fmt.Println("It is sectionvalue")
			fmt.Println(sectionValue)

		}
	}


}


func (c* CfgParser) Get(section string, key string) (string, error) {
	return "", errors.New("hfjh")

}

func isSection(line string) bool {
	match := sectionHeaderRegexp.MatchString(line)
	return match
}


func isKeyValue(line string) bool {
	match := keyValueRegexp.MatchString(line)
	return match
}


