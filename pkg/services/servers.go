package services

import (
	"context"
	"fmt"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	ps "github.com/simonjanss/go-powershell"
)

type ServerService struct {
	db     *gorm.DB
	shell  ps.Shell
	logger *zap.Logger
}

func NewServerService(db *gorm.DB, shell ps.Shell, logger *zap.Logger) ServerService {
	return ServerService{db: db, shell: shell, logger: logger}
}

func (s ServerService) Apply(ctx context.Context, r api.ServerApplyRequest) (*api.ServerApplyResponse, error) {
	logger := s.logger.With(
		zap.String("server", r.Hostname),
		zap.String("os", r.OperatingSystem),
		zap.String("nuix_path", r.NuixPath),
		zap.String("service_account", r.Username),
	)

	if r.OperatingSystem != "linux" && r.OperatingSystem != "windows" {
		logger.Error("specify operating system - 'linux' or 'windows'", zap.String("exception", "invalid operating system"))
		return nil, fmt.Errorf("specify operating_system for %s - 'linux' or 'windows'", r.Hostname)
	}

	// Check if the requested server exists (in that case update it)
	logger.Debug("Checking if server already exists")
	var newSrv api.Server
	if err := s.db.Where("hostname = ?", r.Hostname).First(&newSrv).Error; err != nil {
		// return the error if it isn't a "record not found"-error
		if !gorm.IsRecordNotFoundError(err) {
			logger.Error("Cannot get the server", zap.String("exception", err.Error()))
			return nil, err
		}
		logger.Debug("Server already exist - will update instead of create new")
	}

	// Test connection and install websocket to the server
	// if it is a new server or the nuix-path has been changed
	logger.Debug("Checking if the server should be tested or not")
	if newSrv.ID == 0 || newSrv.NuixPath != r.NuixPath {
		logger.Debug("Testing the server")

		logger.Info("Creating new remote-client for powershell")
		// set options for the connection
		var opts powershell.Options
		opts.Host = newSrv.Hostname
		if len(newSrv.Username) != 0 {
			logger.Debug("Adding credentials for powershell-session")
			opts.Username = newSrv.Username
			opts.Password = newSrv.Password
		}

		// create the client
		client, err := powershell.NewClient(s.shell, opts)
		if err != nil {
			logger.Error("Failed to create remote-client for powershell", zap.String("exception", err.Error()))
			return nil, fmt.Errorf("failed to create remote-client for powershell: %v", err)
		}

		if err := client.CheckPath(r.NuixPath); err != nil {
			logger.Error("Failed to test NuixPath for server", zap.String("exception", err.Error()))
			return nil, fmt.Errorf("failed to test nuix-path for server: %v", err)
		}
	}

	// Set data to the new Server-model
	newSrv.Hostname = r.Hostname
	newSrv.Port = r.Port
	newSrv.Username = r.Username
	newSrv.Password = r.Password
	newSrv.OperatingSystem = r.OperatingSystem
	newSrv.NuixPath = r.NuixPath

	// Save the new NMS to the DB
	logger.Info("Saving server to the DB")
	if err := s.db.Save(&newSrv).Error; err != nil {
		logger.Error("Cannot save server to DB", zap.String("exception", err.Error()))
		return nil, fmt.Errorf("failed to apply server %s : %v", newSrv.Hostname, err)
	}

	logger.Debug("Server has been saved to the DB")
	return &api.ServerApplyResponse{}, nil
}

func (s ServerService) List(ctx context.Context, r api.ServerListRequest) (*api.ServerListResponse, error) {
	s.logger.Debug("Getting Servers-list")
	var servers []api.Server
	if err := s.db.Find(&servers).Error; err != nil {
		s.logger.Error("Cannot get Servers-list", zap.String("exception", err.Error()))
		return nil, err
	}
	s.logger.Debug("Got Servers-list", zap.Int("amount", len(servers)))
	return &api.ServerListResponse{Servers: servers}, nil
}
