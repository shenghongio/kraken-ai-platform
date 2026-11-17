package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/shenghongio/kraken-ai-platform/orchestrator/domain"
)

var ErrInvalidInput = errors.New("invalid input")

// Repository 声明 Service 依赖的任务存储能力
type Repository interface {
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error)
	List(ctx context.Context) ([]*domain.Task, error)
	RequestCancel(ctx context.Context, id domain.TaskID) error
}

// Service 编排任务生命周期相关的应用
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

// CreateInput 创建任务入参数
type CreateInput struct {
	Type           domain.TaskType
	Priority       int
	payload        []byte
	IdempotencyKey string
	MaxRetries     int
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Task, error) {
	if in.Type == "" {
		return nil, fmt.Errorf("task service: type is required: %w", ErrInvalidInput)
	}
	payload := in.payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	task := &domain.Task{
		ID:             domain.NewTaskID(),
		Type:           in.Type,
		Priority:       domain.TaskPriority(in.Priority),
		Payload:        payload,
		IdempotencyKey: in.IdempotencyKey,
		MaxRetries:     in.MaxRetries,
	}
	return s.repo.Create(ctx, task)
}

// List 列出全部任务
func (s *Service) List(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}

// Get 按ID 查询任务
func (s *Service) Get(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task service: id is required: %w", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

// Cancel 请求取消任务
func (s *Service) Cancel(ctx context.Context, id domain.TaskID) error {
	if id == "" {
		return fmt.Errorf("task service: id is required: %w", ErrInvalidInput)
	}
	return s.repo.RequestCancel(ctx, id)
}
