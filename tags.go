package simver

import (
	"context"
	"slices"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/mod/semver"
)

type TagProvider interface {
	FetchTags(ctx context.Context) (Tags, error)
	TagsFromCommit(ctx context.Context, commitHash string) (Tags, error)
	TagsFromBranch(ctx context.Context, branch string) (Tags, error)
}

type TagWriter interface {
	CreateTags(ctx context.Context, tag ...Tag) error
}

type Tag struct {
	Name string
	Ref  string
}

var _ zerolog.LogArrayMarshaler = (*Tags)(nil)

type Tags []Tag

func shortRef(ref string) string {
	if len(ref) <= 11 {
		return ref
	}

	return ref[:4] + "..." + ref[len(ref)-4:]
}

func (t Tags) Sort() Tags {
	tags := make(Tags, len(t))
	copy(tags, t)

	slices.SortFunc(tags, func(a, b Tag) int {
		return semver.Compare(a.Name, b.Name)
	})

	return tags
}

// MarshalZerologArray implements zerolog.LogArrayMarshaler.
func (t Tags) MarshalZerologArray(a *zerolog.Array) {

	tr := t.Sort()

	for _, tag := range tr {
		a.Str(shortRef(tag.Ref) + " => " + tag.Name)
	}
}

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
