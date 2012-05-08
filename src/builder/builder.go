package builder

//ToolWork represents a project that can be downloaded/installed/tested by the
//go tool.
type ToolWork interface {
	Revisions() []string
	ImportPath() string
}

//GopathWork represents a repo that exists as a full GOPATH workspace
type GopathWork interface {
	Revisions() []string
	ClonePath() string
}