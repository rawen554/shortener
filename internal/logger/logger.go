package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("cant create new zap instance: %w", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	return sugar, nil
}
