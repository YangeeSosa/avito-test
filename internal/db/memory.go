package db

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/test-avito/internal/models"
)

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
	ErrPRExists     = errors.New("pull request already exists")
	ErrPRNotFound   = errors.New("pull request not found")
)

type MemoryStore struct {
	mu         sync.RWMutex
	teams      map[string]*models.Team
	users      map[string]*models.User
	prs        map[string]*models.PullRequest
	userPRsIdx map[string]map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	rand.Seed(time.Now().UnixNano())
	return &MemoryStore{
		teams:      make(map[string]*models.Team),
		users:      make(map[string]*models.User),
		prs:        make(map[string]*models.PullRequest),
		userPRsIdx: make(map[string]map[string]struct{}),
	}
}

func (s *MemoryStore) SaveTeam(team models.Team) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.teams[team.Name]; ok {
		return ErrTeamExists
	}

	s.teams[team.Name] = &team
	for _, member := range team.Members {
		s.users[member.UserID] = &models.User{
			ID:       member.UserID,
			Username: member.Username,
			TeamName: team.Name,
			IsActive: member.IsActive,
		}
	}
	return nil
}

func (s *MemoryStore) GetTeam(name string) (*models.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, ok := s.teams[name]
	if !ok {
		return nil, ErrTeamNotFound
	}
	copyTeam := *team
	return &copyTeam, nil
}

func (s *MemoryStore) GetUser(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copyUser := *user
	return &copyUser, nil
}

func (s *MemoryStore) SetUserActive(id string, active bool) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	user.IsActive = active
	copyUser := *user
	return &copyUser, nil
}

func (s *MemoryStore) SavePullRequest(pr models.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.prs[pr.ID]; ok {
		return ErrPRExists
	}
	s.prs[pr.ID] = &pr
	for _, reviewerID := range pr.AssignedReviewers {
		s.addIndexLockHeld(reviewerID, pr.ID)
	}
	return nil
}

func (s *MemoryStore) UpdatePullRequest(pr models.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.prs[pr.ID]
	if !ok {
		return ErrPRNotFound
	}
	for _, oldReviewerID := range stored.AssignedReviewers {
		if idx, ok := s.userPRsIdx[oldReviewerID]; ok {
			delete(idx, pr.ID)
		}
	}
	*stored = pr
	for _, reviewerID := range pr.AssignedReviewers {
		s.addIndexLockHeld(reviewerID, pr.ID)
	}
	return nil
}

func (s *MemoryStore) addIndexLockHeld(userID, prID string) {
	if _, ok := s.userPRsIdx[userID]; !ok {
		s.userPRsIdx[userID] = make(map[string]struct{})
	}
	s.userPRsIdx[userID][prID] = struct{}{}
}

func (s *MemoryStore) GetPullRequest(id string) (*models.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pr, ok := s.prs[id]
	if !ok {
		return nil, ErrPRNotFound
	}
	copyPR := *pr
	copyPR.AssignedReviewers = append([]string(nil), pr.AssignedReviewers...)
	return &copyPR, nil
}

func (s *MemoryStore) ListReviewAssignments(userID string) ([]models.PullRequestShort, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.users[userID]; !ok {
		return nil, ErrUserNotFound
	}

	prIDs := s.userPRsIdx[userID]
	result := make([]models.PullRequestShort, 0, len(prIDs))
	for prID := range prIDs {
		if pr, ok := s.prs[prID]; ok {
			result = append(result, pr.ToShort())
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (s *MemoryStore) SelectRandomActiveMembers(teamName string, exclude map[string]struct{}, limit int) ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, ok := s.teams[teamName]
	if !ok {
		return nil, ErrTeamNotFound
	}
	candidates := make([]*models.User, 0, len(team.Members))
	for _, member := range team.Members {
		if _, skip := exclude[member.UserID]; skip {
			continue
		}
		if !member.IsActive {
			continue
		}
		if user, ok := s.users[member.UserID]; ok && user.IsActive {
			copyUser := *user
			candidates = append(candidates, &copyUser)
		}
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}
