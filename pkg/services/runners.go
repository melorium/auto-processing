package services

import (
	"context"
	"fmt"
	"os"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	avian "github.com/avian-digital-forensics/auto-processing/pkg/avian-client"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
	ps "github.com/simonjanss/go-powershell"

	"github.com/jinzhu/gorm"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RunnerService struct {
	db      *gorm.DB
	shell   ps.Shell
	logger  *zap.Logger
	logPath string
}

func NewRunnerService(db *gorm.DB, shell ps.Shell, logger *zap.Logger, logPath string) RunnerService {
	return RunnerService{db: db, shell: shell, logger: logger, logPath: logPath}
}

// Apply the runner to backend
func (s RunnerService) Apply(ctx context.Context, r api.RunnerApplyRequest) (*api.RunnerApplyResponse, error) {
	s.logger.Debug("Runner apply-request")

	logger := s.logger.With(
		zap.String("runner", r.Name),
		zap.String("hostname", r.Hostname),
		zap.String("nms", r.Nms),
		zap.String("licence", r.Licence),
		zap.Int("workers", int(r.Workers)),
		zap.String("xmx", r.Xmx),
	)

	logger.Debug("Creating runner")
	// Create the requested runner

	// add the switches
	var switches []*api.NuixSwitch
	for _, nuixSwitch := range r.Switches {
		switches = append(switches, &api.NuixSwitch{Value: nuixSwitch})
	}

	runner := api.Runner{
		Name:         r.Name,
		Hostname:     r.Hostname,
		Nms:          r.Nms,
		Licence:      r.Licence,
		Xmx:          r.Xmx,
		Workers:      r.Workers,
		CaseSettings: r.CaseSettings,
		Stages:       r.Stages,
		Switches:     switches,
	}

	// Validate the runner
	logger.Info("Validating runner")
	if err := runner.Validate(); err != nil {
		logger.Error("Validation failed for runner", zap.String("exception", err.Error()))
		return nil, err
	}
	logger.Debug("Validation OK")

	logger.Info("Looking if runner already exists")
	if !s.db.First(&api.Runner{}, "name = ?", runner.Name).RecordNotFound() {
		logger.Error("Create a new runner by a unique name", zap.String("exception", "runner already exists"))
		return nil, fmt.Errorf("runner: %s already exist, create a new runner by a unique name", runner.Name)
	}

	// Check if the requested server exists
	var server api.Server
	logger.Info("Looking if server exists")
	if s.db.First(&server, "hostname = ?", runner.Hostname).RecordNotFound() {
		logger.Error("Requested server for runner does not exist", zap.String("exception", "server not found"))
		return nil, fmt.Errorf("server: %s doesn't exist in the backend, list existing servers by command: 'avian servers list'", runner.Hostname)
	}

	// Check if the requested nms exists
	logger.Info("Looking if NMS exist")
	if s.db.First(&api.Nms{}, "address = ?", runner.Nms).RecordNotFound() {
		logger.Error("Requested NMS for runner does not exist", zap.String("exception", "nms not found"))
		return nil, fmt.Errorf("nms: %s doesn't exist in the backend, list existing nm-servers by command: 'avian nms list'", runner.Nms)
	}

	// Create powershell-connection to test the server
	logger.Info("Creating powershell-session for runner")

	// set options for the connection
	var opts powershell.Options
	opts.Host = server.Hostname
	if len(server.Username) != 0 {
		logger.Debug("Adding credentials for powershell-session")
		opts.Username = server.Username
		opts.Password = server.Password
	}

	// create the client
	client, err := powershell.NewClient(s.shell, opts)
	if err != nil {
		logger.Error("Failed to create remote-client for powershell", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("failed to create remote-client for powershell: %v", err)
	}

	// close the client on exit
	defer client.Close()

	// check that all the paths for the runner exists in the server
	logger.Info("Validating paths for runner")
	for _, path := range runner.Paths() {
		formattedPath := powershell.FormatPath(path)
		if err := client.CheckPath(formattedPath); err != nil {
			logger.Error("Failed to validate path", zap.String("path", path), zap.String("exception", err.Error()))
			return nil, fmt.Errorf("path: %s - err : %v", formattedPath, err)
		}
	}

	// Add the runner to the db
	logger.Info("Saving runner to DB")
	if err := s.db.Save(&runner).Error; err != nil {
		logger.Error("Cannot to save runner to DB", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("failed to create runner: %v", err)
	}

	logger.Info("Runner has been created")
	return &api.RunnerApplyResponse{Runner: runner}, nil
}

func (s RunnerService) List(ctx context.Context, r api.RunnerListRequest) (*api.RunnerListResponse, error) {
	s.logger.Debug("Getting runners-list")
	var runners []api.Runner
	err := s.db.Preload("Stages.Process").
		Preload("Stages.SearchAndTag").
		Preload("Stages.Exclude").
		Preload("Stages.Ocr").
		Preload("Stages.Reload").
		Preload("Stages.Populate").
		Find(&runners).Error
	if err != nil {
		s.logger.Error("Cannot get runners-list", zap.String("exception", err.Error()))
		return nil, err
	}
	s.logger.Debug("Got Runners-list", zap.Int("amount", len(runners)))
	return &api.RunnerListResponse{Runners: runners}, nil
}

func (s RunnerService) Get(ctx context.Context, r api.RunnerGetRequest) (*api.RunnerGetResponse, error) {
	s.logger.Debug("Getting runner", zap.String("runner", r.Name))
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
	if err != nil {
		s.logger.Error("Cannot get runner", zap.String("runner", r.Name), zap.String("exception", err.Error()))
		return nil, err
	}
	s.logger.Debug("Returning runner", zap.String("runner", r.Name))
	return &api.RunnerGetResponse{Runner: runner}, nil
}

func (s RunnerService) StartStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	logger := s.logger.With(zap.String("runner", r.Runner), zap.Int("stage_id", int(r.StageID)))
	logger.Debug("StartStage request")
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		logger.Error("Cannot get the requested stage", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	logger.Debug("Set stage-status to running", zap.Int("stage_id", int(r.StageID)))
	avian.SetStatusRunning(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		logger.Error("Cannot set stage-status to running", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("failed to update stage to running: %v", err)
	}

	logger.Info("STARTING STAGE", zap.String("stage", avian.Name(&stage)))
	return &api.StageResponse{Stage: stage}, nil
}

func (s RunnerService) FailedStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	logger := s.logger.With(zap.String("runner", r.Runner), zap.Int("stage_id", int(r.StageID)))
	logger.Debug("FailedStage request")
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		logger.Error("Cannot get the requested stage", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	logger.Debug("Set stage-status to failed", zap.Int("stage_id", int(r.StageID)))
	avian.SetStatusFailed(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		logger.Error("Cannot set stage-status to failed", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("cannot to update stage to failed: %v", err)
	}

	logger.Info("FAILED STAGE", zap.String("stage", avian.Name(&stage)))
	return &api.StageResponse{Stage: stage}, nil
}

func (s RunnerService) FinishStage(ctx context.Context, r api.StageRequest) (*api.StageResponse, error) {
	logger := s.logger.With(zap.String("runner", r.Runner), zap.Int("stage_id", int(r.StageID)))
	logger.Debug("FinishStage request")
	var stage api.Stage
	if err := s.db.Preload("Process").
		Preload("SearchAndTag").
		Preload("Exclude").
		Preload("Reload").
		Preload("Populate").
		Preload("Ocr").
		First(&stage, r.StageID).Error; err != nil {
		logger.Error("Cannot get the requested stage", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("did not get requested stage : %v", err)
	}

	logger.Debug("Set stage-status to finished", zap.Int("stage_id", int(r.StageID)))
	avian.SetStatusFinished(&stage)
	if err := s.db.Save(&stage).Error; err != nil {
		logger.Error("Cannot set stage-status to finished", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("failed to update stage to running: %v", err)
	}

	logger.Info("FINISHED STAGE", zap.String("stage", avian.Name(&stage)))
	return &api.StageResponse{Stage: stage}, nil
}

// LogItem logs an item that has been processed
func (s RunnerService) LogItem(ctx context.Context, r api.LogItemRequest) (*api.LogResponse, error) {
	logName := fmt.Sprintf("%s%s-item.log", s.logPath, r.Runner)
	logger, err := s.getLogger(logName)
	if err != nil {
		return nil, err
	}

	logger = logger.With(
		zap.String("runner", r.Runner),
		zap.String("stage", r.Stage),
		zap.Int("stage_id", r.StageID),
		zap.Int("count", r.Count),
	)

	if len(r.ProcessStage) > 0 {
		logger = logger.With(zap.String("process_stage", r.ProcessStage))
	}
	if len(r.MimeType) > 0 {
		logger = logger.With(zap.String("mime_type", r.MimeType))
	}
	if len(r.GUID) > 0 {
		logger = logger.With(zap.String("guid", r.GUID))
	}

	logger.Debug(r.Message)
	return &api.LogResponse{}, nil
}

func (s RunnerService) LogDebug(ctx context.Context, r api.LogRequest) (*api.LogResponse, error) {
	logName := fmt.Sprintf("%s%s.log", s.logPath, r.Runner)
	logger, err := s.getLogger(logName)
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("runner", r.Runner))
	if len(r.Stage) > 0 {
		logger = logger.With(zap.Int("stage_id", r.StageID), zap.String("stage", r.Stage))
	}

	logger.Debug(r.Message)
	return &api.LogResponse{}, nil
}

func (s RunnerService) LogInfo(ctx context.Context, r api.LogRequest) (*api.LogResponse, error) {
	logName := fmt.Sprintf("%s%s.log", s.logPath, r.Runner)
	logger, err := s.getLogger(logName)
	if err != nil {
		return nil, err
	}

	logger.With(zap.String("runner", r.Runner))
	if len(r.Stage) > 0 {
		logger = logger.With(zap.Int("stage_id", r.StageID), zap.String("stage", r.Stage))
	}

	logger.Info(r.Message)
	return &api.LogResponse{}, nil
}

func (s RunnerService) LogError(ctx context.Context, r api.LogRequest) (*api.LogResponse, error) {
	logName := fmt.Sprintf("%s%s.log", s.logPath, r.Runner)
	logger, err := s.getLogger(logName)
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("runner", r.Runner))
	if len(r.Stage) > 0 {
		logger = logger.With(zap.Int("stage_id", r.StageID), zap.String("stage", r.Stage))
	}
	if len(r.Exception) > 0 {
		logger = logger.With(zap.String("exception", r.Exception))
	}

	logger.Error(r.Message)
	return &api.LogResponse{}, nil
}

func (s RunnerService) getLogger(logName string) (*zap.Logger, error) {
	log, err := os.OpenFile(logName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log-file %s: %v", logName, err)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   log.Name(),
		MaxSize:    0, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zap.DebugLevel,
	)

	logger := zap.New(core)
	return logger, nil
}
