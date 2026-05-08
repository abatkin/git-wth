package git

type Ref struct {
	SHA  string
	Name string
}

type Git interface {
	Config(key string) (string, error)
	ConfigRegexp(pattern string) ([]string, error)
	Log(format string, revisionRange string) ([]string, error)
	ShowRef() ([]Ref, error)
	SymbolicRefHead() (string, error)
	HasModifiedFiles() (bool, error)
	HasStagedChanges() (bool, error)
}
