package service

import (
	"errors"
	"time"

	"github.com/test-avito/internal/db"
	"github.com/test-avito/internal/models"
)

var (
	ErrPRMerged        = errors.New("pull request already merged")
	ErrReviewerMissing = errors.New("reviewer not assigned to PR")
	ErrNoCandidate     = errors.New("no active candidate available")
)

type Service struct {
	store *db.MemoryStore
}

func New(store *db.MemoryStore) *Service {
	return &Service{store: store}
}

func (s *Service) CreateTeam(team models.Team) (*models.Team, error) {
	if err := s.store.SaveTeam(team); err != nil {
		return nil, err
	}
	saved, _ := s.store.GetTeam(team.Name)
	return saved, nil
}

func (s *Service) GetTeam(name string) (*models.Team, error) {
	return s.store.GetTeam(name)
}

func (s *Service) SetUserActive(userID string, active bool) (*models.User, error) {
	user, err := s.store.SetUserActive(userID, active)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) CreatePullRequest(pr models.PullRequest) (*models.PullRequest, error) {
	author, err := s.store.GetUser(pr.AuthorID)
	if err != nil {
		return nil, err
	}

	exclude := map[string]struct{}{pr.AuthorID: {}}
	reviewers, err := s.store.SelectRandomActiveMembers(author.TeamName, exclude, 2)
	if err != nil && err != db.ErrTeamNotFound {
		return nil, err
	}

	assigned := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		assigned = append(assigned, reviewer.ID)
	}

	pr.Status = models.PRStatusOpen
	pr.AssignedReviewers = assigned
	now := time.Now().UTC()
	pr.CreatedAt = &now

	if err := s.store.SavePullRequest(pr); err != nil {
		return nil, err
	}
	saved, _ := s.store.GetPullRequest(pr.ID)
	return saved, nil
}

func (s *Service) MergePullRequest(id string) (*models.PullRequest, error) {
	pr, err := s.store.GetPullRequest(id)
	if err != nil {
		return nil, err
	}
	if pr.Status == models.PRStatusMerged {
		return pr, nil
	}
	pr.Status = models.PRStatusMerged
	now := time.Now().UTC()
	pr.MergedAt = &now
	if err := s.store.UpdatePullRequest(*pr); err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) ReassignReviewer(prID, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := s.store.GetPullRequest(prID)
	if err != nil {
		return nil, "", err
	}
	if pr.Status == models.PRStatusMerged {
		return nil, "", ErrPRMerged
	}
	idx := -1
	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, "", ErrReviewerMissing
	}

	reviewer, err := s.store.GetUser(oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	exclude := map[string]struct{}{oldReviewerID: {}}
	for _, r := range pr.AssignedReviewers {
		exclude[r] = struct{}{}
	}

	candidates, err := s.store.SelectRandomActiveMembers(reviewer.TeamName, exclude, 1)
	if err != nil {
		return nil, "", err
	}
	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}
	replacement := candidates[0]
	pr.AssignedReviewers[idx] = replacement.ID

	if err := s.store.UpdatePullRequest(*pr); err != nil {
		return nil, "", err
	}
	return pr, replacement.ID, nil
}

func (s *Service) ListUserReviews(userID string) ([]models.PullRequestShort, error) {
	return s.store.ListReviewAssignments(userID)
}
