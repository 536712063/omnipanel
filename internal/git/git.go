package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Repo struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	RemoteURL string `json:"remote_url,omitempty"`
}

type CloneRequest struct {
	URL        string `json:"url"`
	LocalPath  string `json:"local_path"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	SSHKeyPath string `json:"ssh_key_path,omitempty"`
}

type Branch struct {
	Name      string    `json:"name"`
	IsCurrent bool      `json:"is_current"`
	IsRemote  bool      `json:"is_remote"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Commit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Date      time.Time `json:"date"`
}

type StatusItem struct {
	File   string `json:"file"`
	Status string `json:"status"`
}

func Open(repoPath string) (*Repo, error) {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}
	return buildRepo(repoPath, repo)
}

func Clone(req CloneRequest) (*Repo, error) {
	opts := &gogit.CloneOptions{
		URL:      req.URL,
		Progress: os.Stdout,
	}
	if req.Username != "" {
		opts.Auth = &http.BasicAuth{Username: req.Username, Password: req.Password}
	}

	repo, err := gogit.PlainClone(req.LocalPath, false, opts)
	if err != nil {
		return nil, fmt.Errorf("clone repo: %w", err)
	}
	return buildRepo(req.LocalPath, repo)
}

func buildRepo(path string, repo *gogit.Repository) (*Repo, error) {
	name := filepath.Base(path)
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	remoteURL := ""
	for _, r := range remotes {
		if r.Config().Name == "origin" {
			urls := r.Config().URLs
			if len(urls) > 0 {
				remoteURL = urls[0]
			}
			break
		}
	}
	return &Repo{
		ID:        path,
		Path:      path,
		Name:      name,
		RemoteURL: remoteURL,
	}, nil
}

func (r *Repo) Branches() ([]Branch, error) {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return nil, err
	}

	refIter, err := repo.References()
	if err != nil {
		return nil, err
	}
	defer refIter.Close()

	var branches []Branch
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	currentBranch := head.Name().Short()

	err = refIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, Branch{
				Name:      ref.Name().Short(),
				IsCurrent: ref.Name().Short() == currentBranch,
				IsRemote:  false,
				Hash:      ref.Hash().String(),
			})
		} else if ref.Name().IsRemote() {
			branches = append(branches, Branch{
				Name:     ref.Name().Short(),
				IsRemote: true,
				Hash:     ref.Hash().String(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })
	return branches, nil
}

func (r *Repo) Log(limit int) ([]Commit, error) {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return nil, err
	}
	logIter, err := repo.Log(&gogit.LogOptions{Order: gogit.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}
	defer logIter.Close()

	var commits []Commit
	for i := 0; i < limit; i++ {
		c, err := logIter.Next()
		if err != nil {
			break
		}
		commits = append(commits, Commit{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:7],
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Message:   c.Message,
			Date:      c.Author.When,
		})
	}
	return commits, nil
}

func (r *Repo) Status() ([]StatusItem, error) {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return nil, err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	var items []StatusItem
	for file, s := range status {
		var statusStr string
		switch {
		case s.Staging == gogit.Untracked && s.Worktree == gogit.Untracked:
			statusStr = "untracked"
		case s.Staging == gogit.Modified || s.Worktree == gogit.Modified:
			statusStr = "modified"
		case s.Staging == gogit.Added || s.Worktree == gogit.Added:
			statusStr = "added"
		case s.Staging == gogit.Deleted || s.Worktree == gogit.Deleted:
			statusStr = "deleted"
		case s.Staging == gogit.Renamed || s.Worktree == gogit.Renamed:
			statusStr = "renamed"
		case s.Staging == gogit.Copied || s.Worktree == gogit.Copied:
			statusStr = "copied"
		default:
			statusStr = fmt.Sprintf("%v %v", s.Staging, s.Worktree)
		}
		items = append(items, StatusItem{File: file, Status: statusStr})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].File < items[j].File })
	return items, nil
}

func (r *Repo) Pull(ctx context.Context) error {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Pull(&gogit.PullOptions{RemoteName: "origin"})
}

func (r *Repo) Push(ctx context.Context) error {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return err
	}
	return repo.Push(&gogit.PushOptions{})
}

func (r *Repo) Commit(message string) (string, error) {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	if err := wt.AddGlob("."); err != nil {
		return "", err
	}
	commit, err := wt.Commit(message, &gogit.CommitOptions{})
	if err != nil {
		return "", err
	}
	return commit.String(), nil
}

func (r *Repo) Checkout(branchName string) error {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
	})
}

func (r *Repo) CreateBranch(name string) error {
	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + name),
		Create: true,
	})
}
