package simver

import (
	"strings"

	"golang.org/x/mod/semver"
)

type Tag struct {
	Name string
	Ref  string
}

type Tags []Tag

func (t Tags) GetReserved() (Tag, bool) {

	for _, tag := range t {
		if strings.Contains(tag.Name, "-reserved") {
			return tag, true
		}
	}

	return Tag{}, false
}

func (t Tags) Names() []string {
	var names []string

	for _, tag := range t {
		names = append(names, tag.Name)
	}

	return names
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

func (t Tags) ExtractCommitRefs() Tags {
	var tags Tags

	for _, tag := range t {
		if len(tag.Ref) == 40 {
			tags = append(tags, tag)
		}
	}

	return tags
}

func (t Tags) MappedByName() map[string]string {
	m := make(map[string]string)

	for _, tag := range t {
		m[tag.Name] = tag.Ref
	}

	return m
}
