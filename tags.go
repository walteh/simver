package simver

import (
	"errors"
	"regexp"
	"sort"

	"golang.org/x/mod/semver"
)

type TagInfo struct {
	Name string
	Ref  string
}

type Tags []TagInfo

// HighestSemverContainingString finds the highest semantic version tag that contains the specified string.
func (t Tags) HighestSemverMatching(matcher ...*regexp.Regexp) (string, error) {
	var versions []string

	for _, m := range matcher {
		for _, tag := range t {
			if m.MatchString(tag.Name) {
				// Attempt to parse the semantic version from the tag
				v := semver.Canonical(tag.Name)
				if v != "" && semver.IsValid(v) {
					versions = append(versions, tag.Name)
				}
			}
		}
		if len(versions) == 0 {
			return "", errors.New("no matching semantic versions found")
		}
	}
	// Use sort to find the highest version
	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare(versions[i], versions[j]) < 0
	})

	return versions[len(versions)-1], nil
}
