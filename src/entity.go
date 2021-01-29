package upgradebot

type Commit struct {
	Sha string
}

type File struct {
	Status    string `json:"status"`
	Filename  string `json:"filename"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
}

func (f *File) getTotalModifications() int {
	return f.Changes + f.Deletions + f.Additions
}

type PullRequestData struct {
	Number   int    `json:"number"`
	HtmlUrl  string `json:"html_url"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Comments int    `json:"comments"`
	ClosedAt string `json:"closed_at"`
}

type PullRequest struct {
	Data  PullRequestData
	Files []File
}

type CommitChanges struct {
	Commits []Commit
	Files   []File
}

type ReleaseData struct {
	Name        string
	Body        string
	Prerelease  bool
	Tag         string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
}

type TagCompare struct {
	PullRequests []PullRequest
	Files        []File
}

type CreatePullRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft"`
}
