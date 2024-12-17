package automaxprocs

import (
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/forest33/mqtt-sync/pkg/logger"

	"github.com/forest33/mqtt-sync/business/entity"
)

func Init(cfg *entity.Config, log *logger.Logger) error {
	if cfg.Runtime.GoMaxProcs != 0 {
		return nil
	}

	undo, err := maxprocs.Set(maxprocs.Logger(log.Printf))
	defer undo()

	return err
}
