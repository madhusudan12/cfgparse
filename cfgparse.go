package cfgparse

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

var (
	sectionHeaderRegexp = regexp.MustCompile("\\[([^]]+)\\]")
	keyValueRegexp = regexp.MustCompile("([^:=\\s][^:=]*)\\s*(?P<vi>[:=])\\s*(.*)$")
)

var allowedTypes = []string{".ini", ".cfg"}

type section struct {
	name string
	items map[string]interface{}
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


func isValidType(fileType string) bool {
	for _, value := range allowedTypes {
		if value == fileType {
			return true
		}
	}
	return false
}


func getFileType(filename string) (string, error) {
	fileType := filepath.Ext(filename)
	if ! isValidType(fileType) {
		errMessage := "File type not supported. Supported types (" + strings.Join(allowedTypes, " ") + ")"
		err := errors.New(errMessage)
		return fileType, err
	}
	return fileType, nil
}


func (c* CfgParser)setDelimitor() {
	switch c.fileType {
	case ".ini":
		c.delimeter = "="
		break
	case ".cfg":
		c.delimeter = ":"
		break
	default:
		c.delimeter = ":"
	}
}


func (c* CfgParser) ReadFile(filename string) error {
	if len(filename) == 0 {
		err := errors.New("File name cannot be empty")
		return err
	}
	fileType, err := getFileType(filename)
	if err != nil {
		return err
	}
	c.fileType = fileType
	c.setDelimitor()
	cfgFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	c.Parse(cfgFile)
	return nil
}


func getKeyValuefromSectionValue(sectionValue string, sep string, lineNo int)(string, string) {
	defer func() {
		err := recover()
		if err != nil {
			errMessage := fmt.Sprintf("Config file format error at line no %d. Please format it correctly",lineNo)
			panic(errMessage)
		}
	}()
	keyValues := strings.Split(sectionValue, sep)
	key := keyValues[0]
	value := keyValues[1]
	return key, value
}


func (c* CfgParser) Parse(cfgFile *os.File) {
	reader := bufio.NewReader(cfgFile)
	var lineNo int
	var curSection section
	var err error

	for err == nil {
		buff, _, err := reader.ReadLine()
		if err != nil{
			break
		}
		if len(buff) == 0 {
			continue
		}
		line := strings.TrimFunc(string(buff), unicode.IsSpace)
		lineNo++
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if isSection(line) {
			sectionHeader := sectionHeaderRegexp.FindStringSubmatch(line)[1]
			curSection = section{}
			curSection.name = sectionHeader
			curSection.items = make(map[string]interface{})
			// TODO: check for dulicate sections
			if c.sections == nil {
				c.sections = make(map[string]section)
			}
			c.sections[curSection.name] = curSection
		} else if isKeyValue(line) {
			sectionValue := keyValueRegexp.FindStringSubmatch(line)[0]
			fmt.Println("keyvalue", sectionValue)
			key, value := getKeyValuefromSectionValue(sectionValue, c.delimeter, lineNo)
			curSection.items[key] = value
		}
	}
}

func (c* CfgParser) GetAllSections() []string{
	sections := []string{}
	for section := range c.sections {
		sections = append(sections, section)
	}
	return sections
}

func (c* CfgParser) Get(section string, key string) interface{} {
	sectionValue, ok := c.sections[section]
	if !ok {
		errMessage := fmt.Sprintf("No such section %s exists", section)
		panic(errMessage)
	}
	value, ok := sectionValue.items[key]
	if !ok {
		errMessage := fmt.Sprintf("No such key %s exists in section %s exists", key, section)
		panic(errMessage)
	}
	return value
}


func isSection(line string) bool {
	match := sectionHeaderRegexp.MatchString(line)
	return match
}


func isKeyValue(line string) bool {
	match := keyValueRegexp.MatchString(line)
	return match
}


