package simver

import "context"

type CommitInfo struct {
	Hash         string
	Message      string
	IsPrMerge    bool
	AssociatedPR int
}

type GitProvider interface {
	FetchTags(ctx context.Context) ([]TagInfo, error)
	GetHeadRef(ctx context.Context) (string, error)
	CreateTag(ctx context.Context, tag TagInfo) error
	CommitFromRef(ctx context.Context, ref string) (string, error)
	Branch(ctx context.Context) (string, error)
	TagsFromCommit(ctx context.Context, commitHash string) ([]TagInfo, error)
	TagsFromBranch(ctx context.Context, branch string) ([]TagInfo, error)
}
