package gitexec

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/go-faster/errors"
	"github.com/rs/zerolog"
	"github.com/walteh/simver"
)

var (
	ErrExecGit = errors.New("simver.ErrExecGit")
)

var _ simver.GitProvider = (*gitProvider)(nil)

type gitProvider struct {
	RepoPath      string
	Token         string
	User          string
	Email         string
	TokenEnvName  string
	GitExecutable string
	ReadOnly      bool
	Org           string
	Repo          string
}

type GitProviderOpts struct {
	RepoPath      string
	Token         string
	User          string
	Email         string
	TokenEnvName  string
	GitExecutable string
	ReadOnly      bool
	Org           string
	Repo          string
}

func (p *gitProvider) RepoName(_ context.Context) (string, string, error) {
	return p.Org, p.Repo, nil
}

// func NewLocalReadOnlyGitProvider(executable string, repoPath string) (simver.GitProvider, error) {
// 	return &gitProvider{
// 		RepoPath:      repoPath,
// 		Token:         "",
// 		TokenEnvName:  "",
// 		User:          "",
// 		Email:         "",
// 		GitExecutable: executable,
// 		ReadOnly:      true,
// 	}, nil
// }

// func NewLocalReadOnlyTagProvider(executable string, repoPath string) (simver.TagProvider, error) {
// 	return &gitProvider{
// 		RepoPath:      repoPath,
// 		Token:         "",
// 		TokenEnvName:  "",
// 		User:          "",
// 		Email:         "",
// 		GitExecutable: executable,
// 		ReadOnly:      true,
// 	}, nil
// }

func NewGitProvider(opts *GitProviderOpts) (*gitProvider, error) {
	if opts.RepoPath == "" {
		return nil, errors.Wrap(ErrExecGit, "repo path is required")
	}

	if !opts.ReadOnly && opts.Token == "" {
		return nil, errors.Wrap(ErrExecGit, "token is required for read/write")
	}

	if opts.User == "" {
		return nil, errors.Wrap(ErrExecGit, "user is required")
	}

	if opts.Email == "" {
		return nil, errors.Wrap(ErrExecGit, "email is required")
	}

	if opts.TokenEnvName == "" {
		return nil, errors.Wrap(ErrExecGit, "token env name is required")
	}

	if opts.GitExecutable == "" {
		opts.GitExecutable = "git"
	}

	if opts.Org == "" {
		return nil, errors.Wrap(ErrExecGit, "org is required")
	}

	if opts.Repo == "" {
		return nil, errors.Wrap(ErrExecGit, "repo is required")
	}

	// check if git is in PATH
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.Wrap(ErrExecGit, "git executable is required")
	}

	return &gitProvider{
		RepoPath:      opts.RepoPath,
		Token:         opts.Token,
		User:          opts.User,
		Email:         opts.Email,
		TokenEnvName:  opts.TokenEnvName,
		GitExecutable: opts.GitExecutable,
		ReadOnly:      opts.ReadOnly,
		Org:           opts.Org,
		Repo:          opts.Repo,
	}, nil
}

func (p *gitProvider) git(ctx context.Context, str ...string) *exec.Cmd {
	env := []string{
		p.TokenEnvName + "=" + p.Token,
		"COMMITTER_NAME=" + p.User,
		"COMMITTER_EMAIL=" + p.Email,
		"AUTHOR_NAME=" + p.User,
		"AUTHOR_EMAIL=" + p.Email,
	}

	if len(str) > 0 && str[0] == "git" {
		// If the first argument is git, remove it because it's already in the command and will never be valid
		str = str[1:]
	}

	zerolog.Ctx(ctx).Debug().Strs("args", str).Str("executable", p.GitExecutable).Msg("building git command")

	cmd := exec.CommandContext(ctx, p.GitExecutable, str...)
	cmd.Dir = p.RepoPath
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)

	return cmd
}

func (p *gitProvider) CommitFromRef(ctx context.Context, str string) (string, error) {

	zerolog.Ctx(ctx).Debug().Str("ref", str).Msg("getting commit from ref")

	cmd := p.git(ctx, "rev-parse", str)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "git rev-parse "+str)
	}

	res := strings.TrimSpace(string(out))

	zerolog.Ctx(ctx).Debug().Str("ref", str).Str("commit", res).Msg("got commit from ref")

	return res, nil
}

func (p *gitProvider) Branch(ctx context.Context) (string, error) {

	zerolog.Ctx(ctx).Debug().Msg("getting branch")

	cmd := p.git(ctx, "branch", "--contains", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "git branch --contains HEAD")
	}
	lines := strings.Split(string(out), "\n")
	res := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			res = strings.TrimPrefix(line, "* ")
			break
		}
	}

	if res == "" {
		return "", errors.Wrap(ErrExecGit, "could not find current branch")
	}

	zerolog.Ctx(ctx).Debug().Str("branch", res).Msg("got branch")

	return res, nil
}

func (p *gitProvider) GetHeadRef(ctx context.Context) (string, error) {

	zerolog.Ctx(ctx).Debug().Msg("getting head ref")

	// Get the current HEAD ref
	cmd := p.git(ctx, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "git rev-parse HEAD")
	}

	res := strings.TrimSpace(string(out))

	zerolog.Ctx(ctx).Debug().Str("ref", res).Msg("got head ref")

	return res, nil
}
