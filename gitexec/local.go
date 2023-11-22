package gitexec

import (
	"context"
	"strings"

	"github.com/go-faster/errors"
	"github.com/spf13/afero"
	"github.com/walteh/simver"
)

func BuildLocalProviders(fls afero.Fs) (simver.GitProvider, simver.TagReader, simver.TagWriter, simver.PRResolver, error) {

	repoData, err := fls.Open(".git/config")
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "error opening /.git/config")
	}

	repoConfig, err := afero.ReadAll(repoData)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "error reading /.git/config")
	}

	// split the config file into lines
	lines := strings.Split(string(repoConfig), "\n")

	// find the remote origin section
	var remoteOrigin []string
	for i, line := range lines {
		if strings.HasPrefix(line, "[remote \"origin\"]") {
			remoteOrigin = lines[i:]
			break
		}

	}

	if len(remoteOrigin) == 0 {
		return nil, nil, nil, nil, errors.New("could not find remote origin in /.git/config")
	}

	// find the url line
	var urlLine string
	for _, line := range remoteOrigin {
		if strings.HasPrefix(line, "	url = ") {
			urlLine = line
			break
		}
	}

	if urlLine == "" {
		return nil, nil, nil, nil, errors.New("could not find url line in remote origin in /.git/config")
	}

	// grab the url
	url := strings.TrimSpace(strings.TrimPrefix(urlLine, "url = "))

	// split the url into parts
	parts := strings.Split(url, "/")

	// grab the org and repo
	org := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	var path string
	if bp, ok := fls.(*afero.BasePathFs); ok {
		path, err = bp.RealPath("/")
		if err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "error getting real path")
		}
	} else {
		path = "."
	}

	c := &GitProviderOpts{
		RepoPath:      path,
		Token:         "invalid",
		User:          "invalid",
		Email:         "invalid",
		TokenEnvName:  "GITHUB_TOKEN",
		GitExecutable: "git",
		ReadOnly:      true,
		Org:           org,
		Repo:          repo,
	}

	git, err := NewGitProvider(c)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "error creating git provider")
	}

	return git, git, git, &LocalPullRequestResolver{}, nil
}

type LocalPullRequestResolver struct {
	git simver.GitProvider
}

func (p *LocalPullRequestResolver) CurrentPR(ctx context.Context) (*simver.PRDetails, error) {

	cmt, err := p.git.GetHeadRef(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting head ref")
	}

	return &simver.PRDetails{
		Number:               1,
		HeadBranch:           "local",
		BaseBranch:           "local",
		RootBranch:           "local",
		Merged:               false,
		MergeCommit:          cmt,
		HeadCommit:           cmt,
		PotentialMergeCommit: cmt,
		BaseCommit:           cmt,
		RootCommit:           cmt,
	}, nil
}
