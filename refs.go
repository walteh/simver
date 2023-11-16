package simver

type RefProvider interface {
	Head() string
	Base() string
	Root() string
	Merge() string
}

type BasicRefProvider struct {
	HeadRef  string
	BaseRef  string
	RootRef  string
	MergeRef string
}

func (e *BasicRefProvider) Head() string {
	return e.HeadRef
}

func (e *BasicRefProvider) Base() string {
	return e.BaseRef
}

func (e *BasicRefProvider) Root() string {
	return e.RootRef
}

func (e *BasicRefProvider) Merge() string {
	return e.MergeRef
}
