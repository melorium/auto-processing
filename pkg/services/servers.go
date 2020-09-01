package services

import (
	"context"
	"fmt"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"

	"github.com/jinzhu/gorm"
	ps "github.com/simonjanss/go-powershell"
)

type ServerService struct {
	db    *gorm.DB
	shell ps.Shell
}

func NewServerService(db *gorm.DB, shell ps.Shell) ServerService {
	return ServerService{db: db, shell: shell}
}

func (s ServerService) Apply(ctx context.Context, r api.ServerApplyRequest) (*api.ServerApplyResponse, error) {
	if r.OperatingSystem != "linux" && r.OperatingSystem != "windows" {
		return nil, fmt.Errorf("specify operating_system for %s - 'linux' or 'windows'", r.Hostname)
	}

	// Check if the requested server exists (in that case update it)
	var newSrv api.Server
	if err := s.db.Where("hostname = ?", r.Hostname).First(&newSrv).Error; err != nil {
		// return the error if it isn't a "record not found"-error
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}

	// Test connection and install websocket to the server
	// if it is a new server or the nuix-path has been changed
	if (newSrv.ID == 0 || newSrv.NuixPath != r.NuixPath) && !r.SkipInstall {
		client, err := powershell.NewClient(r.Hostname, s.shell)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote-client for powershell: %v", err)
		}

		nuixPath := powershell.FormatPath(r.NuixPath)
		if err := client.SetupNuix(nuixPath); err != nil {
			return nil, err
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
	if err := s.db.Save(&newSrv).Error; err != nil {
		return nil, fmt.Errorf("failed to apply server %s : %v", newSrv.Hostname, err)
	}

	return &api.ServerApplyResponse{}, nil
}

func (s ServerService) List(ctx context.Context, r api.ServerListRequest) (*api.ServerListResponse, error) {
	var servers []api.Server
	err := s.db.Find(&servers).Error
	return &api.ServerListResponse{Servers: servers}, err
}
