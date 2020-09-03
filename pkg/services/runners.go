package services

import (
	"context"
	"fmt"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	avian "github.com/avian-digital-forensics/auto-processing/pkg/avian-client"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
	ps "github.com/simonjanss/go-powershell"

	"github.com/jinzhu/gorm"
)

type RunnerService struct {
	db    *gorm.DB
	shell ps.Shell
}

func NewRunnerService(db *gorm.DB, shell ps.Shell) RunnerService {
	return RunnerService{db: db, shell: shell}
}

// Apply the runner to backend
func (s RunnerService) Apply(ctx context.Context, r api.RunnerApplyRequest) (*api.RunnerApplyResponse, error) {
	// Create the requested runner
	runner := api.Runner{
		Name:         r.Name,
		Hostname:     r.Hostname,
		Nms:          r.Nms,
		Licence:      r.Licence,
		Xmx:          r.Xmx,
		Workers:      r.Workers,
		CaseSettings: r.CaseSettings,
		Stages:       r.Stages,
	}

	// Validate the runner
	if err := runner.Validate(); err != nil {
		return nil, err
	}

	if !s.db.First(&api.Runner{}, "name = ?", runner.Name).RecordNotFound() {
		return nil, fmt.Errorf("runner: %s already exist, create a new runner by a unique name", runner.Name)
	}

	// Check if the requested server exists
	if s.db.First(&api.Server{}, "hostname = ?", runner.Hostname).RecordNotFound() {
		return nil, fmt.Errorf("server: %s doesn't exist in the backend, list existing servers by command: 'avian servers list'", runner.Hostname)
	}

	// Check if the requested nms exists
	if s.db.First(&api.Nms{}, "address = ?", runner.Nms).RecordNotFound() {
		return nil, fmt.Errorf("nms: %s doesn't exist in the backend, list existing nm-servers by command: 'avian nms list'", runner.Nms)
	}

	// Create powershell-connection to test the server
	client, err := powershell.NewClient(runner.Hostname, s.shell)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote-client for powershell: %v", err)
	}
	defer client.Close()

	// check that all the paths for the runner exists in the server
	for _, path := range runner.Paths() {
		formattedPath := powershell.FormatPath(path)
		if err := client.CheckPath(formattedPath); err != nil {
			return nil, fmt.Errorf("path: %s - err : %v", formattedPath, err)
		}
	}

	// Add the runner to the db
	if err := s.db.Save(&runner).Error; err != nil {
		return nil, fmt.Errorf("failed to create runner: %v", err)
	}

	return &api.RunnerApplyResponse{Runner: runner}, nil
}

func (s RunnerService) List(ctx context.Context, r api.RunnerListRequest) (*api.RunnerListResponse, error) {
	var runners []api.Runner
	err := s.db.Preload("Stages.Process").
		Preload("Stages.SearchAndTag").
		Preload("Stages.Exclude").
		Preload("Stages.Ocr").
		Preload("Stages.Reload").
		Preload("Stages.Populate").
		Find(&runners).Error
	return &api.RunnerListResponse{Runners: runners}, err
}

func (s RunnerService) Get(ctx context.Context, r api.RunnerGetRequest) (*api.RunnerGetResponse, error) {
	var runner api.Runner
	err := s.db.Preload("Stages.Process").
		Preload("Stages.SearchAndTag").
		Preload("Stages.Exclude").
		Preload("Stages.Ocr").
		Preload("Stages.Reload").
		Preload("Stages.Populate").
		Preload("CaseSettings.Case").
		Preload("CaseSettings.CompoundCase").
		Preload("CaseSettings.ReviewCompound").
		First(&runner, "name = ?", r.Name).Error
	return &api.RunnerGetResponse{Runner: runner}, err
}

func (s RunnerService) StartStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	avian.SetStatusRunning(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		return nil, fmt.Errorf("failed to update stage to running: %v", err)
	}

	return &api.StageResponse{Stage: stage}, nil
}

func (s RunnerService) FailedStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	avian.SetStatusFailed(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		return nil, fmt.Errorf("failed to update stage to running: %v", err)
	}

	return &api.StageResponse{Stage: stage}, nil
}

func (s RunnerService) FinishStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	avian.SetStatusFinished(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		return nil, fmt.Errorf("failed to update stage to running: %v", err)
	}

	return &api.StageResponse{Stage: stage}, nil
}
