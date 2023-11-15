package simver

import "context"

type GitProvider interface {
	GetHeadRef(ctx context.Context) (string, error)
	CommitFromRef(ctx context.Context, ref string) (string, error)
	Branch(ctx context.Context) (string, error)
}

type TagProvider interface {
	FetchTags(ctx context.Context) ([]Tag, error)
	CreateTag(ctx context.Context, tag Tag) error
	TagsFromCommit(ctx context.Context, commitHash string) ([]Tag, error)
	TagsFromBranch(ctx context.Context, branch string) ([]Tag, error)
}
