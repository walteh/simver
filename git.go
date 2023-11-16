package simver

import "context"

type GitProvider interface {
	GetHeadRef(ctx context.Context) (string, error)
	CommitFromRef(ctx context.Context, ref string) (string, error)
	Branch(ctx context.Context) (string, error)
}
