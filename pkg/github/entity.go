package github

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

func (f *File) GetTotalModifications() int {
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
	Title  string                `json:"title"`
	Body   string                `json:"body"`
	Head   CreatePullRequestHead `json:"head"`
	Base   CreatePullRequestBase `json:"base"`
	Draft  bool                  `json:"draft"`
	Labels []string              `json:"labels"`
}

type CreatePullRequestHead struct {
	Label string `json:"label"`
}

type CreatePullRequestBase struct {
	Ref string `json:"ref"`
}

type LabelsRequestData []LabelRequestData

type LabelRequestData struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

type LabelsRequest struct {
	Labels []string `json:"labels"`
}
