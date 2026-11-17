package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shenghongio/kraken-ai-platform/orchestrator/domain"
)

var ErrInvalidInput = errors.New("invalid input")

// ClaimResult 是任务领取任务成功后返回给Agent的结果
type ClaimResult struct {
	Task       *domain.Task
	Attempt    int
	LeaseEpoch int64
}

// Repository 声明Service依赖的Agent 能力
type Repository interface {
	UpdateAgent(ctx context.Context, agent *domain.Agent) error
	Heartbeat(ctx context.Context, id string, busySlots int) error
}

// TaskRepository 声明 Service 依赖的任务领取能力
type TaskRepository interface {
	ClaimTask(ctx context.Context, agentID string) (*ClaimResult, error)
}

// Service 编排 Agent 注册，心跳与任务领取
type Service struct {
	repo     Repository
	taskRepo TaskRepository
	now      func() time.Time
}

func NewService(repo Repository, taskRepo TaskRepository) *Service {
	return &Service{repo: repo, taskRepo: taskRepo, now: func() time.Time { return time.Now().UTC() }}
}

// RegisterInput 注册 Agent 的入参
type RegisterInput struct {
	AgentID        string
	Version        string
	Capabilities   []string
	MaxConcurrency int
}

// Register 校验并注册 (或更新)Agent
func (s *Service) Register(ctx context.Context, in RegisterInput) (*domain.Agent, error) {
	if in.AgentID == "" {
		return nil, fmt.Errorf("agent service: agent_id is required: %w", ErrInvalidInput)
	}
	if in.MaxConcurrency <= 0 {
		return nil, fmt.Errorf("agent service: max_concurrency must be positive > 0: %w", ErrInvalidInput)
	}
	agent := &domain.Agent{
		ID:             in.AgentID,
		Status:         domain.AgentStatusOnline,
		Version:        in.Version,
		Capabilities:   in.Capabilities,
		MaxConcurrency: in.MaxConcurrency,
		LastHeartbeat:  s.now(),
	}
	if err := s.repo.UpdateAgent(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}
