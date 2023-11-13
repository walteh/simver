package simver

import "context"

type GitProvider interface {
	GetHeadRef(ctx context.Context) (string, error)
	CommitFromRef(ctx context.Context, ref string) (string, error)
	Branch(ctx context.Context) (string, error)
}

type TagProvider interface {
	FetchTags(ctx context.Context) ([]TagInfo, error)
	CreateTag(ctx context.Context, tag TagInfo) error
	TagsFromCommit(ctx context.Context, commitHash string) ([]TagInfo, error)
	TagsFromBranch(ctx context.Context, branch string) ([]TagInfo, error)
}
