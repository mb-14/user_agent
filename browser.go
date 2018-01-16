// Copyright (C) 2012-2018 Miquel Sabaté Solà <mikisabate@gmail.com>
// This file is licensed under the MIT license.
// See the LICENSE file.

package user_agent

import (
	"regexp"
	"strings"
)

var ie11Regexp = regexp.MustCompile("^rv:(.+)$")

// A struct containing all the information that we might be
// interested from the browser.
type Browser struct {
	// The name of the browser's engine.
	Engine string `json:"engine"`

	// The version of the browser's engine.
	EngineVersion string `json:"engine_version"`

	// The name of the browser.
	Name string `json:"name"`

	// The version of the browser.
	Version string `json:"version"`
}

// Extract all the information that we can get from the User-Agent string
// about the browser and update the receiver with this information.
//
// The function receives just one argument "sections", that contains the
// sections from the User-Agent string after being parsed.
func (p *UserAgent) detectBrowser(sections []section) {
	slen := len(sections)

	if sections[0].name == "Opera" {
		p.Browser.Name = "Opera"
		p.Browser.Version = sections[0].version
		p.Browser.Engine = "Presto"
		if slen > 1 {
			p.Browser.EngineVersion = sections[1].version
		}
	} else if sections[0].name == "Dalvik" {
		// When Dalvik VM is in use, there is no browser info attached to ua.
		// Although browser is still a Mozilla/5.0 compatible.
		p.Mozilla = "5.0"
	} else if slen > 1 {
		engine := sections[1]
		p.Browser.Engine = engine.name
		p.Browser.EngineVersion = engine.version
		if slen > 2 {
			sectionIndex := 2
			// The version after the engine comment is empty on e.g. Ubuntu
			// platforms so if this is the case, let's use the next in line.
			if sections[2].version == "" && slen > 3 {
				sectionIndex = 3
			}
			p.Browser.Version = sections[sectionIndex].version
			if engine.name == "AppleWebKit" {
				switch sections[slen-1].name {
				case "Edge":
					p.Browser.Name = "Edge"
					p.Browser.Version = sections[slen-1].version
					p.Browser.Engine = "EdgeHTML"
					p.Browser.EngineVersion = ""
				case "OPR":
					p.Browser.Name = "Opera"
					p.Browser.Version = sections[slen-1].version
				default:
					if sections[sectionIndex].name == "Chrome" {
						p.Browser.Name = "Chrome"
					} else if sections[sectionIndex].name == "Chromium" {
						p.Browser.Name = "Chromium"
					} else {
						p.Browser.Name = "Safari"
					}
				}
			} else if engine.name == "Gecko" {
				name := sections[2].name
				if name == "MRA" && slen > 4 {
					name = sections[4].name
					p.Browser.Version = sections[4].version
				}
				p.Browser.Name = name
			} else if engine.name == "like" && sections[2].name == "Gecko" {
				// This is the new user agent from Internet Explorer 11.
				p.Browser.Engine = "Trident"
				p.Browser.Name = "Internet Explorer"
				for _, c := range sections[0].comment {
					version := ie11Regexp.FindStringSubmatch(c)
					if len(version) > 0 {
						p.Browser.Version = version[1]
						return
					}
				}
				p.Browser.Version = ""
			}
		}
	} else if slen == 1 && len(sections[0].comment) > 1 {
		comment := sections[0].comment
		if comment[0] == "compatible" && strings.HasPrefix(comment[1], "MSIE") {
			p.Browser.Engine = "Trident"
			p.Browser.Name = "Internet Explorer"
			// The MSIE version may be reported as the compatibility version.
			// For IE 8 through 10, the Trident token is more accurate.
			// http://msdn.microsoft.com/en-us/library/ie/ms537503(v=vs.85).aspx#VerToken
			for _, v := range comment {
				if strings.HasPrefix(v, "Trident/") {
					switch v[8:] {
					case "4.0":
						p.Browser.Version = "8.0"
					case "5.0":
						p.Browser.Version = "9.0"
					case "6.0":
						p.Browser.Version = "10.0"
					}
					break
				}
			}
			// If the Trident token is not provided, fall back to MSIE token.
			if p.Browser.Version == "" {
				p.Browser.Version = strings.TrimSpace(comment[1][4:])
			}
		}
	}
}
