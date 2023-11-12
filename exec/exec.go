package exec

type ExecProvider struct {
	RepoPath      string
	Token         string
	User          string
	Email         string
	TokenEnvName  string
	GitExecutable string
}
