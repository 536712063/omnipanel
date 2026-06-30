package git

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu    sync.RWMutex
	repos map[string]*Repo
}

func NewService() *Service {
	return &Service{
		repos: make(map[string]*Repo),
	}
}

func (s *Service) AddLocalRepo(path string) (*Repo, error) {
	repo, err := Open(path)
	if err != nil {
		return nil, err
	}
	repo.ID = uuid.New().String()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.repos[repo.ID] = repo
	return repo, nil
}

func (s *Service) CloneRepo(req CloneRequest) (*Repo, error) {
	repo, err := Clone(req)
	if err != nil {
		return nil, err
	}
	repo.ID = uuid.New().String()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.repos[repo.ID] = repo
	return repo, nil
}

func (s *Service) GetRepo(id string) (*Repo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repo, ok := s.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo not found")
	}
	return repo, nil
}

func (s *Service) ListRepos() []Repo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Repo, 0, len(s.repos))
	for _, repo := range s.repos {
		result = append(result, *repo)
	}
	return result
}

func (s *Service) RemoveRepo(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.repos, id)
}

func (s *Service) Branches(repoID string) ([]Branch, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return nil, err
	}
	return repo.Branches()
}

func (s *Service) Log(repoID string, limit int) ([]Commit, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return repo.Log(limit)
}

func (s *Service) Status(repoID string) ([]StatusItem, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return nil, err
	}
	return repo.Status()
}

func (s *Service) Pull(ctx context.Context, repoID string) error {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return err
	}
	return repo.Pull(ctx)
}

func (s *Service) Push(ctx context.Context, repoID string) error {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return err
	}
	return repo.Push(ctx)
}

func (s *Service) Commit(repoID string, message string) (string, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return "", err
	}
	return repo.Commit(message)
}

func (s *Service) Checkout(repoID string, branchName string) error {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return err
	}
	return repo.Checkout(branchName)
}

func (s *Service) CreateBranch(repoID string, name string) error {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return err
	}
	return repo.CreateBranch(name)
}

func (s *Service) RecentActivity(repoID string, limit int) ([]Commit, error) {
	return s.Log(repoID, limit)
}

func (s *Service) GetRepoStats(repoID string) (map[string]interface{}, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return nil, err
	}
	commits, err := repo.Log(1000)
	if err != nil {
		return nil, err
	}
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	status, err := repo.Status()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"commit_count":    len(commits),
		"branch_count":    len(branches),
		"uncommitted_count": len(status),
		"last_commit_at": func() string {
			if len(commits) > 0 {
				return commits[0].Date.Format(time.RFC3339)
			}
			return ""
		}(),
	}, nil
}
