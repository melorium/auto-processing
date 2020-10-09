package heartbeat

import (
	"time"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	"github.com/avian-digital-forensics/auto-processing/pkg/avian-client"
	"github.com/avian-digital-forensics/auto-processing/pkg/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Service struct {
	runnersvc services.RunnerService
	pause     time.Duration
	db        *gorm.DB
	logger    *zap.Logger
}

func New(r services.RunnerService, logger *zap.Logger) Service {
	return Service{r, 2 * time.Minute, r.DB, logger}
}

func (s Service) Beat() {
	for {
		var runners []api.Runner
		var lastCheck = time.Now().Add(-s.pause)
		var query = s.db.Where("active = ? AND healthy_at < ?", true, lastCheck)
		if err := query.Find(&runners).Error; err != nil {
			s.logger.Error("Failed to fetch runners", zap.String("exception", err.Error()))
		}
		s.logger.Info("Got unhealthy runners from db", zap.Int("amount", len(runners)))

		for _, runner := range runners {
			runner.Status = avian.StatusTimeout
			runner.Active = false
			if err := s.db.Save(&runner).Error; err != nil {
				s.logger.Error("Cannot save the failed runner", zap.String("exception", err.Error()))
			}

			// Set servers activity
			if err := s.runnersvc.SetServerActivity(runner, false); err != nil {
				s.logger.Error("Cannot save the failed runner", zap.String("exception", err.Error()))
			}

			// update nms information
			if err := s.runnersvc.ResetNms(runner); err != nil {
				s.logger.Error("Cannot save the failed runner", zap.String("exception", err.Error()))
			}

			if err := s.runnersvc.RemoveScript(runner); err != nil {
				s.logger.Error("Cannot remove script for runner", zap.String("exception", err.Error()))
			}
		}

		time.Sleep(s.pause)
	}
}
