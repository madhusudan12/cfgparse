package cfgparse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var (
	sectionRegexp = regexp.MustCompile("\\[([^]]+)\\]")
	keyValueRegexp = regexp.MustCompile("([^:=\\s][^:=]*)\\s*(?P<vi>[:=])\\s*(.*)$")
	interpolateRegexp = regexp.MustCompile("%\\(([^)]*)\\)s|.")
)
const MaxDepth = 10
var allowedTypes = []string{".ini", ".cfg"}

type section struct {
	name  string
	filePosition int64
	items map[string]string
}

type CfgParser struct {
	fileName  string
	fileType  string
	sections  map[string]section
	delimeter string
	mutex     sync.Mutex
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
	if !isValidType(fileType) {
		errMessage := "File type not supported. Supported types (" + strings.Join(allowedTypes, " ") + ")"
		err := errors.New(errMessage)
		return fileType, err
	}
	return fileType, nil
}

func (c *CfgParser) setDelimitor() {
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

func (c *CfgParser) ReadFile(fileName string) error {
	if len(fileName) == 0 {
		err := errors.New("file name cannot be empty")
		return err
	}
	fileType, err := getFileType(fileName)
	c.fileName = fileName
	if err != nil {
		return err
	}
	c.fileType = fileType
	c.setDelimitor()
	cfgFile, err := os.Open(fileName)
	defer cfgFile.Close()
	if err != nil {
		return err
	}
	c.Parse(cfgFile)
	return nil
}

func getKeyValuefromSectionValue(sectionValue string, sep string, lineNo uint) (string, string) {
	defer func() {
		err := recover()
		if err != nil {
			errMessage := fmt.Sprintf("Config file format error at line no %d. Please format it correctly", lineNo)
			panic(errMessage)
		}
	}()
	keyValues := strings.Split(sectionValue, sep)
	key := keyValues[0]
	value := keyValues[1]
	return key, value
}

func (c *CfgParser) Parse(cfgFile *os.File) {
	reader := bufio.NewReader(cfgFile)
	var lineNo uint
	var curSection section
	var filePos int64
	var numOfBytes int
	for {
		buff, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		if len(buff) == 0 {
			filePos++
			continue
		}
		numOfBytes = len(buff)
		filePos = filePos + int64(numOfBytes) + 1
		line := strings.TrimFunc(string(buff), unicode.IsSpace)
		lineNo++
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if isSection(line) {
			sectionHeader := sectionRegexp.FindStringSubmatch(line)[1]
			curSection = section{}
			if c.isSectionAlreadyExists(sectionHeader) {
				errMessage := fmt.Sprintf("Parsing Error: Duplicate section %s occured at line %d",sectionHeader, lineNo)
				panic(errMessage)
			}
			curSection.name = sectionHeader
			curSection.items = make(map[string]string)
			curSection.filePosition = filePos
			if c.sections == nil {
				c.sections = make(map[string]section)
			}
			c.sections[curSection.name] = curSection
		} else if isKeyValue(line) {
			sectionValue := keyValueRegexp.FindStringSubmatch(line)[0]
			key, value := getKeyValuefromSectionValue(sectionValue, c.delimeter, lineNo)
			pos := strings.Index(";", value)         // Checking for comments
			if pos > -1 {
				if v := value[pos-1]; unicode.IsSpace(rune(v)) {
					value = value[:pos-1]
				}
			}
			curSection.items[key] = value
		}
	}
}

func (c *CfgParser) GetAllSections() []string {
	sections := []string{}
	for section := range c.sections {
		sections = append(sections, section)
	}
	return sections
}

func (c *CfgParser) Items(section string) map[string]string {
	sectionValue, ok := c.sections[section]
	if !ok {
		errMessage := fmt.Sprintf("No such section %s exists", section)
		panic(errMessage)
	}
	return sectionValue.items
}

func (c *CfgParser) Get(sectionName string, key string) string {
	sectionValue, ok := c.sections[sectionName]
	if !ok {
		errMessage := fmt.Sprintf("No such section %s exists", sectionName)
		panic(errMessage)
	}
	value, ok := sectionValue.items[key]
	if !ok {
		errMessage := fmt.Sprintf("No such key %s exists in section %s", key, sectionName)
		panic(errMessage)
	}
	return c.interpolate(sectionName, key, value)
}

func (c *CfgParser) GetBool(section string, key string) (bool, error) {
	value := c.Get(section, key)
	if resValue, err := strconv.ParseBool(value); err != nil {
		return resValue, nil
	} else {
		ErrMessage := fmt.Sprintf("Cannot convert %s to type bool", value)
		err := errors.New(ErrMessage)
		return resValue, err
	}
}

func (c *CfgParser) GetInt(section string, key string) (int64, error) {
	value := c.Get(section, key)
	if resValue, err := strconv.Atoi(value); err != nil {
		return int64(resValue), nil
	} else {
		ErrMessage := fmt.Sprintf("Cannot convert %s to type int64", value)
		err := errors.New(ErrMessage)
		return int64(resValue), err
	}
}

func (c *CfgParser) GetFloat(section string, key string) (float64, error) {
	value := c.Get(section, key)
	if resValue, err := strconv.ParseFloat(value, 64); err != nil {
		return resValue, nil
	} else {
		ErrMessage := fmt.Sprintf("Cannot convert %s to type float64", value)
		err := errors.New(ErrMessage)
		return resValue, err
	}
}

func (c *CfgParser) AddSection(sectionName string) error {
	newSection := section{}
	if c.isSectionAlreadyExists(sectionName) {
		errMessage := fmt.Sprintf("Cannot add section %s already exits", sectionName)
		err := errors.New(errMessage)
		return err
	}
	c.mutex.Lock()
	newSection.name = sectionName
	newSection.items = make(map[string]string)
	if c.sections == nil {
		c.sections = make(map[string]section)
	}
	f, err := os.OpenFile(c.fileName, os.O_APPEND|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		errMesssage := fmt.Sprintf("Somthing went wrong while opening file %s. Check if is opened in other places", c.fileName)
		err = errors.New(errMesssage)
		return err
	}
	writer := bufio.NewWriter(f)
	//TODO: add two new lines only if last char in file is not '\n'
	buff := "\n\n[" + sectionName + "]\n"
	fileStat, err := f.Stat()
	if err != nil {
		errMesssage := fmt.Sprintf("Somthing went wrong while opening file %s. Check if is opened in other places", c.fileName)
		err = errors.New(errMesssage)
		return err
	}
	filePosition := fileStat.Size()
	newSection.filePosition = filePosition + int64(len(buff))
	c.sections[newSection.name] = newSection
	_, writerErr := writer.WriteString(buff)
	if writerErr != nil {
		errMesssage := fmt.Sprintf("Somthing went wrong while writing into file %s. Check if is opened in other places", c.fileName)
		err = errors.New(errMesssage)
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	c.mutex.Unlock()
	return nil
}

// TODO: Apply locks while writing also load data in memory after writing into file, along with updating the file positions
// TODO: find the best method to update filepositions ( reload the entire file , change all file positions greater than the current writing section
func (c *CfgParser) Set(sectionName string, key string, value string) {
	if !c.isSectionAlreadyExists(sectionName) {
		err := c.AddSection(sectionName)
		if err != nil {
			panic("Error adding section name")
		}
	}
	filePos, err := c.getSectionPos(sectionName)
	fReader, err := os.OpenFile(c.fileName, os.O_RDONLY, 0644)
	if err != nil {
		panic("Error accessing the config file")
	}
	defer fReader.Close()
	fileStat, err := fReader.Stat()
	if err != nil {
		panic("Error accessing the config file")
	}
	fileSize := fileStat.Size()
	sectionPositon, err := fReader.Seek(int64(filePos), 0)
	if err != nil {
		panic("Error accessing the config file")
	}
	extraFileSize := fileSize - sectionPositon + 1
	buffBytes := make([]byte, extraFileSize)
	_ , err = fReader.ReadAt(buffBytes, sectionPositon)
	var remainingSlice string
	if err != io.EOF {
		errMessage := fmt.Sprintf("Error Reading the config file %v", err)
		panic(errMessage)
	}
	if len(buffBytes) == 0 {
		remainingSlice = ""
	} else {
		remainingSlice = string(buffBytes)[:len(buffBytes)-1]
	}
	keyValueToWrite := key + c.delimeter + value
	dataToWrite := keyValueToWrite + "\n" + remainingSlice
	bytesToWrite := []byte(dataToWrite)
	c.mutex.Lock()
	fWriter, err := os.OpenFile(c.fileName, os.O_WRONLY, 0644)
	if err != nil {
		panic("Error accessing the config file")
	}
	bytesAdded , wErr := fWriter.WriteAt(bytesToWrite, sectionPositon)
	if wErr != nil {
		errMsg := fmt.Sprintf("Error Writing to config file %v", wErr)
		panic(errMsg)
	}
	c.sections[sectionName].items[key] = value
	fWriter.Close()
	noOfExtraBytes := bytesAdded - len(remainingSlice)
	c.reOrderFilePositions(sectionPositon, noOfExtraBytes)
	c.mutex.Unlock()
}


func (c *CfgParser) reOrderFilePositions(sectionPosition int64, bytesAdded int) {
	for sec, secObj := range c.sections {
		if secObj.filePosition > sectionPosition {
			secObj.filePosition = c.sections[sec].filePosition + int64(bytesAdded)
			c.sections[sec] = secObj
		}
	}
}

func (c *CfgParser) interpolate(sectionName string, key string, value string) string {
	for depth := 0; depth < MaxDepth; depth++ {
		if strings.Contains(value,"%(") {
			value = interpolateRegexp.ReplaceAllStringFunc(value, func(m string) string {
				match := interpolateRegexp.FindAllStringSubmatch(m, 1)[0][1]
				replacement := c.Get(sectionName, match)
				return replacement
			})
		}
	}
	return value
}

func (c *CfgParser) getSectionPos(sectionName string) (int64, error){
	for sec, _ := range c.sections {
		if sec == sectionName {
			return c.sections[sectionName].filePosition, nil
		}
	}
	return 0, errors.New("No section exists")
}

func isSection(line string) bool {
	match := sectionRegexp.MatchString(line)
	return match
}

func (c *CfgParser) isSectionAlreadyExists(sectionName string) bool {
	for section, _ := range c.sections {
		if section == sectionName {
			return true
		}
	}
	return false
}

func isKeyValue(line string) bool {
	match := keyValueRegexp.MatchString(line)
	return match
}
