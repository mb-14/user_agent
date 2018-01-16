// Copyright (C) 2012-2018 Miquel Sabaté Solà <mikisabate@gmail.com>
// This file is licensed under the MIT license.
// See the LICENSE file.

// Package user_agent implements an HTTP User Agent string parser. It defines
// the type UserAgent that contains all the information from the parsed string.
// It also implements the Parse function and getters for all the relevant
// information that has been extracted from a parsed User Agent string.
package user_agent

import (
	"encoding/json"
	"strings"
)

// A section contains the name of the product, its version and
// an optional comment.
type section struct {
	name    string
	version string
	comment []string
}

// The UserAgent struct contains all the info that can be extracted
// from the User-Agent string.
type UserAgent struct {
	ua           string  `json:"-"`
	Mozilla      string  `json:"mozilla"`
	Platform     string  `json:"platform"`
	Os           string  `json:"os"`
	Localization string  `json:"localization"`
	Browser      Browser `json:"browser"`
	Bot          bool    `json:"is_bot"`
	Mobile       bool    `json:"is_mobile"`
	Undecided    bool    `json:"-"`
}

// Read from the given string until the given delimiter or the
// end of the string have been reached.
//
// The first argument is the user agent string being parsed. The second
// argument is a reference pointing to the current index of the user agent
// string. The delimiter argument specifies which character is the delimiter
// and the cat argument determines whether nested '(' should be ignored or not.
//
// Returns an array of bytes containing what has been read.
func readUntil(ua string, index *int, delimiter byte, cat bool) []byte {
	var buffer []byte

	i := *index
	catalan := 0
	for ; i < len(ua); i = i + 1 {
		if ua[i] == delimiter {
			if catalan == 0 {
				*index = i + 1
				return buffer
			}
			catalan--
		} else if cat && ua[i] == '(' {
			catalan++
		}
		buffer = append(buffer, ua[i])
	}
	*index = i + 1
	return buffer
}

// Parse the given product, that is, just a name or a string
// formatted as Name/Version.
//
// It returns two strings. The first string is the name of the product and the
// second string contains the version of the product.
func parseProduct(product []byte) (string, string) {
	prod := strings.SplitN(string(product), "/", 2)
	if len(prod) == 2 {
		return prod[0], prod[1]
	}
	return string(product), ""
}

// Parse a section. A section is typically formatted as follows
// "Name/Version (comment)". Both, the comment and the version are optional.
//
// The first argument is the user agent string being parsed. The second
// argument is a reference pointing to the current index of the user agent
// string.
//
// Returns a section containing the information that we could extract
// from the last parsed section.
func parseSection(ua string, index *int) (s section) {
	buffer := readUntil(ua, index, ' ', false)

	s.name, s.version = parseProduct(buffer)
	if *index < len(ua) && ua[*index] == '(' {
		*index++
		buffer = readUntil(ua, index, ')', true)
		s.comment = strings.Split(string(buffer), "; ")
		*index++
	}
	return s
}

// Initialize the parser.
func (p *UserAgent) initialize() {
	p.ua = ""
	p.Mozilla = ""
	p.Platform = ""
	p.Os = ""
	p.Localization = ""
	p.Browser.Engine = ""
	p.Browser.EngineVersion = ""
	p.Browser.Name = ""
	p.Browser.Version = ""
	p.Bot = false
	p.Mobile = false
	p.Undecided = false
}

// Parse the given User-Agent string and get the resulting UserAgent object.
//
// Returns an UserAgent object that has been initialized after parsing
// the given User-Agent string.
func New(ua string) *UserAgent {
	o := &UserAgent{}
	o.Parse(ua)
	return o
}

// Parse the given User-Agent string. After calling this function, the
// receiver will be setted up with all the information that we've extracted.
func (p *UserAgent) Parse(ua string) {
	var sections []section

	p.initialize()
	p.ua = ua
	for index, limit := 0, len(ua); index < limit; {
		s := parseSection(ua, &index)
		if !p.Mobile && s.name == "Mobile" {
			p.Mobile = true
		}
		sections = append(sections, s)
	}

	if len(sections) > 0 {
		if sections[0].name == "Mozilla" {
			p.Mozilla = sections[0].version
		}

		p.detectBrowser(sections)
		p.detectOS(sections[0])

		if p.Undecided {
			p.checkBot(sections)
		}
	}
}

func (ua *UserAgent) Map() map[string]interface{} {
	jsonString, _ := json.Marshal(ua)
	var uaMap map[string]interface{}
	json.Unmarshal(jsonString, &uaMap)
	return uaMap
}
