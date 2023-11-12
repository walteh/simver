package main

import (
	"context"
	"os"
	"time"

	"github.com/walteh/simver"
	"github.com/walteh/simver/exec"
)

func NewGitHubActionsProvider() (simver.GitProvider, simver.PRProvider, error) {

	token := os.Getenv("GITHUB_TOKEN")
	repoPath := os.Getenv("GITHUB_WORKSPACE")

	c := &exec.ExecProvider{
		RepoPath:      repoPath,
		Token:         token,
		User:          "github-actions[bot]",
		Email:         "41898282+github-actions[bot]@users.noreply.github.com",
		TokenEnvName:  "GITHUB_TOKEN",
		GitExecutable: "git",
	}

	pr := &exec.ExecGHProvider{
		GitHubToken:  token,
		RepoPath:     repoPath,
		GHExecutable: "gh",
	}

	return c, pr, nil
}

func main() {

	ctx := context.Background()

	gitprov, prprov, err := NewGitHubActionsProvider()
	if err != nil {
		panic(err)
	}

	ee, err := simver.LoadExecution(ctx, gitprov, prprov)
	if err != nil {
		panic(err)
	}

	tags := simver.NewTags(ee)

	reservedTag, reserved := tags.GetReserved()

	for !reserved {
		err := gitprov.CreateTag(ctx, reservedTag)
		if err != nil {
			time.Sleep(1 * time.Second)
			ee, err := simver.LoadExecution(ctx, gitprov, prprov)
			if err != nil {
				panic(err)
			}
			tags := simver.NewTags(ee)
			reservedTag, reserved = tags.GetReserved()
		} else {
			reserved = true
		}
	}

	for _, tag := range tags {
		if tag.Name == reservedTag.Name && tag.Ref == reservedTag.Ref {
			continue
		}

		err := gitprov.CreateTag(ctx, tag)
		if err != nil {
			panic(err)
		}
	}

}
