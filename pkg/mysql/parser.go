package mysql

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

func NewConfigParser() (t *ConfigParser) {
	t = new(ConfigParser)
	t.Data = make(map[string]*ConfigSection)
	return t
}

type ConfigParser struct {
	Data map[string]*ConfigSection
	lock sync.Mutex
}

func (t *ConfigParser) Parse(r io.Reader) (err error) {
	sectionReg := regexp.MustCompile(`^\[`)
	commentReg := regexp.MustCompile(`^#`)

	scanner := bufio.NewScanner(r)
	var currentSection *ConfigSection

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if string(line) == "" {
			continue
		}

		if sectionReg.Match(line) {
			sectionName := strings.Trim(strings.Trim(string(line), "["), "]")
			if currentSection, _ = t.GetSection(sectionName); currentSection == nil {
				currentSection = NewConfigSection(sectionName)
				t.Data[sectionName] = currentSection
			}
		} else if commentReg.Match(line) {

		} else {
			arr := strings.Split(string(line), "=")
			currentSection.Set(arr[0], strings.Join(arr[1:], "="))
		}
	}
	return err
}

func (t *ConfigParser) ParseFile(configFile string) (err error) {
	f, err := os.OpenFile(configFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Parse(f)
}

func (t *ConfigParser) DeleteSection(sectionName string, section *ConfigSection) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.Data, sectionName)
}

func (t *ConfigParser) SetSection(section *ConfigSection) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Data[section.Name] = section
}

func (t *ConfigParser) GetSection(sectionName string) (section *ConfigSection, err error) {
	if section, ok := t.Data[sectionName]; ok {
		return section, nil
	}
	return nil, fmt.Errorf("section not found")
}

func (t *ConfigParser) String() (s string) {
	for _, v := range t.Data {
		s += "[" + v.Name + "]\n"
		s += v.String() + "\n"
	}
	return s
}

func NewConfigSection(name string) (t *ConfigSection) {
	t = new(ConfigSection)
	t.Data = make(map[string]string)
	t.Name = name
	return
}

type ConfigSection struct {
	Name string
	Data map[string]string
	lock sync.Mutex
}

func (t *ConfigSection) Set(key, value string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Data[key] = value
}

func (t *ConfigSection) Get(key string) (s string, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if s, ok := t.Data[key]; ok {
		return s, nil
	}
	return s, fmt.Errorf("key not found")
}

func (t *ConfigSection) Delete(key string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.Data, key)
}

func (t *ConfigSection) String() (s string) {
	reg := regexp.MustCompile(`^\d+$`)
	for k, v := range t.Data {
		if reg.Match([]byte(v)) {
			s += k + "=" + v + "\n"
		} else {
			s += k + `="` + v + `"` + "\n"
		}
	}
	return s
}
