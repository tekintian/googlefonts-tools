package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)
const (
	LineEndings = "\n" // OS-specific line endings
)

type IniFile struct {
	fileName      string
	caseSensitive bool
	sections      map[string][]string
	sectionNames  []string
	modified      bool
}

func NewIniFile(filename string, casesensitive bool) *IniFile {
	ini := &IniFile{
		fileName:      filename,
		caseSensitive: casesensitive,
		sections:      make(map[string][]string),
		modified:      false,
	}
	ini.loadValues()
	return ini
}

func (ini *IniFile) FileName() string {
	return ini.fileName
}

func (ini *IniFile) CaseSensitive() bool {
	return ini.caseSensitive
}

func (ini *IniFile) String() string {
	var buf bytes.Buffer
	for _, sec := range ini.sectionNames {
		if lst, ok := ini.sections[sec]; ok {
			buf.WriteString(fmt.Sprintf("[%s]", sec) + LineEndings)
			for _, s := range lst {
				buf.WriteString(fmt.Sprintf("%s", s) + LineEndings)
			}
			buf.WriteString(LineEndings)
		}
	}

	return strings.TrimRight(buf.String(), LineEndings)
}

func (ini *IniFile) getRealValue(s string) string {
	if !ini.caseSensitive {
		return strings.ToLower(s)
	}
	return s
}

func (ini *IniFile) loadValues() {
	if !FileExists(ini.fileName) {
		return
	}

	file, err := os.Open(ini.fileName)
	if err != nil {
		return
	}
	defer file.Close()

	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		s = strings.TrimSpace(s)
		s = ini.getRealValue(s)
		if s != "" && s[0] != ';' {
			if s[0] == '[' && s[len(s)-1] == ']' {
				s = s[1 : len(s)-1]
				section = s
				ini.sectionNames = append(ini.sectionNames, s)
			} else {
				if section != "" {
					if pos := strings.Index(s, "="); pos > 0 {
						if sl, ok := ini.sections[section]; ok {
							ini.sections[section] = append(sl, s)
						} else {
							ini.sections[section] = []string{s}
						}
					} else {
						// ingore invalid ident
						//
					}
				}
			}
		}
	}
}

func (ini *IniFile) flush() {
	file, err := os.Create(ini.fileName)
	if err != nil {
		return
	}
	defer file.Close()

	fw := bufio.NewWriter(file)
	_, err = fw.WriteString(ini.String())
	if err != nil {
		return
	}
	ini.modified = (fw.Flush() != nil)
}

func (ini *IniFile) SectionExists(section string) bool {
	sec := ini.getRealValue(section)
	if _, ok := ini.sections[sec]; ok {
		return true
	}
	return false
}

func (ini *IniFile) ReadSections() []string {
	var ss []string
	for sec, _ := range ini.sections {
		ss = append(ss, sec)
	}
	return ss
}

func (ini *IniFile) EraseSection(section string) {
	sec := ini.getRealValue(section)
	if _, ok := ini.sections[sec]; ok {
		delete(ini.sections, sec)
		for i, s := range ini.sectionNames {
			if s == sec {
				var ss []string
				for j := 0; j < i; j++ {
					ss = append(ss, ini.sectionNames[j])
				}
				for j := i + 1; j < len(ini.sectionNames); j++ {
					ss = append(ss, ini.sectionNames[j])
				}
				ini.sectionNames = ss
				ini.modified = true
				return
			}
		}
	}
}

func (ini *IniFile) ReadSectionIdents(section string) []string {
	var ss []string
	sec := ini.getRealValue(section)
	if sl, ok := ini.sections[sec]; ok {
		for _, s := range sl {
			if pos := strings.Index(s, "="); pos > 0 {
				ss = append(ss, s[0:pos])
			}
		}
	}
	return ss
}

func (ini *IniFile) ReadSectionValues(section string) []string {
	var ss []string
	sec := ini.getRealValue(section)
	if sl, ok := ini.sections[sec]; ok {
		for _, s := range sl {
			ss = append(ss, s)
		}
	}
	return ss
}

func (ini *IniFile) DeleteIdent(section, ident string) {
	sec := ini.getRealValue(section)
	id := ini.getRealValue(ident)
	if sl, ok := ini.sections[sec]; ok {
		for i := 0; i < len(sl); i++ {
			s := sl[i]
			if pos := strings.Index(s, "="); pos > 0 {
				if s[0:pos] == id {
					var ss []string
					for j := 0; j < i; j++ {
						ss = append(ss, sl[j])
					}
					for j := i + 1; j < len(sl); j++ {
						ss = append(ss, sl[j])
					}
					ini.sections[sec] = ss
					ini.modified = true
					return
				}
			}
		}
	}
}

func (ini *IniFile) IdentExists(section, ident string) bool {
	sec := ini.getRealValue(section)
	id := ini.getRealValue(ident)
	if sl, ok := ini.sections[sec]; ok {
		for _, s := range sl {
			if pos := strings.Index(s, "="); pos > 0 {
				if s[0:pos] == id {
					return true
				}
			}
		}
	}
	return false
}

func (ini *IniFile) ReadString(section, ident, defaultValue string) string {
	sec := ini.getRealValue(section)
	id := ini.getRealValue(ident)
	if sl, ok := ini.sections[sec]; ok {
		for _, s := range sl {
			if pos := strings.Index(s, "="); pos > 0 {
				if s[0:pos] == id {
					return s[pos+1:]
				}
			}
		}
	}
	return defaultValue
}

func (ini *IniFile) WriteString(section, ident, value string) {
	sec := ini.getRealValue(section)
	id := ini.getRealValue(ident)
	if sl, ok := ini.sections[sec]; ok {
		for i := 0; i < len(sl); i++ {
			s := sl[i]
			if pos := strings.Index(s, "="); pos > 0 {
				if s[0:pos] == id {
					var ss []string
					for j := 0; j < i; j++ {
						ss = append(ss, sl[j])
					}
					ss = append(ss, ident+"="+value)
					for j := i + 1; j < len(sl); j++ {
						ss = append(ss, sl[j])
					}
					ini.sections[sec] = ss
					ini.modified = true
					return
				}
			}
		}
		ini.sections[sec] = append(sl, ident+"="+value)
		ini.modified = true
	} else {
		ini.sectionNames = append(ini.sectionNames, sec)
		ini.sections[sec] = []string{ident + "=" + value}
		ini.modified = true
	}
}

func (ini *IniFile) ReadInt(section, ident string, defaultValue int) int {
	s := ini.ReadString(section, ident, "")
	if ret, err := strconv.Atoi(s); err == nil {
		return ret
	} else {
		return defaultValue
	}
}

func (ini *IniFile) WriteInt(section, ident string, value int) {
	ini.WriteString(section, ident, strconv.Itoa(value))
}

func (ini *IniFile) ReadBool(section, ident string, defaultValue bool) bool {
	s := ini.ReadString(section, ident, BoolToStr(defaultValue))
	return StrToBool(s)
}

func (ini *IniFile) WriteBool(section, ident string, value bool) {
	ini.WriteString(section, ident, BoolToStr(value))
}

func (ini *IniFile) ReadFloat(section, ident string, defaultValue float64) float64 {
	s := ini.ReadString(section, ident, "")
	if s != "" {
		if ret, err := strconv.ParseFloat(s, 64); err == nil {
			return ret
		}
	}
	return defaultValue
}

func (ini *IniFile) WriteFloat(section, ident string, value float64) {
	ini.WriteString(section, ident, FloatToStr(value))
}

func (ini *IniFile) Close() {
	if ini.modified {
		ini.flush()
	}
	ini.clear()
}

func (ini *IniFile) clear() {
	ini.fileName = ""
	ini.caseSensitive = false
	ini.sections = make(map[string][]string)
	ini.sectionNames = nil
	ini.modified = false
}

func IniReadString(fileName, section, ident, defaultValue string) string {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.ReadString(section, ident, defaultValue)
}

func IniWriteString(fileName, section, ident, value string) {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	inifile.WriteString(section, ident, value)
}

func IniReadInt(fileName, section, ident string, defaultValue int) int {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.ReadInt(section, ident, defaultValue)
}

func IniWriteInt(fileName, section, ident string, value int) {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	inifile.WriteInt(section, ident, value)
}

func IniSectionExists(fileName, section string) bool {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.SectionExists(section)
}

func IniReadSectionIdents(fileName, section string) []string {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.ReadSectionIdents(section)
}

func IniReadSections(fileName string) []string {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.ReadSections()
}

func IniReadSectionValues(fileName, section string) []string {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.ReadSectionValues(section)
}

func IniEraseSection(fileName, section string) {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	inifile.EraseSection(section)
}

func IniIdentExists(fileName, section, ident string) bool {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	return inifile.IdentExists(section, ident)
}

func IniDeleteIdent(fileName, section, ident string) {
	inifile := NewIniFile(fileName, false)
	defer inifile.Close()
	inifile.DeleteIdent(section, ident)
}
