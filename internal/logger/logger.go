package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config encapsulates the configurable elements of the logger.
type Config struct {
	Level  *zapcore.Level `json:"level,omitempty"  usage:"adjust the verbosity of the logs"`
	Format string         `json:"format,omitempty" usage:"configure the format of the logs"`
}

// Setup creates a logger given the provided configuration.
func Setup(cfg Config) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level.SetLevel(*(cfg.Level))
	zapConfig.Encoding = cfg.Format
	return zapConfig.Build()
}
