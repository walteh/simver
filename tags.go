package simver

import (
	"strings"

	"golang.org/x/mod/semver"
)

type TagInfo struct {
	Name string
	Ref  string
}

type Tags []TagInfo

func (t Tags) GetReserved() (TagInfo, bool) {

	for _, tag := range t {
		if strings.Contains(tag.Name, "-reserved") {
			return tag, true
		}
	}

	return TagInfo{}, false
}

// HighestSemverContainingString finds the highest semantic version tag that contains the specified string.
func (t Tags) HighestSemverMatching(matches []string) (string, error) {

	// matches := t.SemversMatching(func(tag string) bool {
	// 	for _, m := range matcher {
	// 		if m.MatchString(tag) {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// })

	semver.Sort(matches)

	return matches[len(matches)-1], nil
}

func (t Tags) SemversMatching(matcher func(string) bool) []string {
	var versions []string

	for _, tag := range t {
		if matcher(tag.Name) {
			// Attempt to parse the semantic version from the tag
			v := semver.Canonical(tag.Name)
			if v != "" && semver.IsValid(v) {
				versions = append(versions, tag.Name)
			}
		}
	}

	semver.Sort(versions)

	return versions
}
