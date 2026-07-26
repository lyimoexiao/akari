package yggdrasiladapter

import "go.uber.org/zap"

type SigningFailureReporter struct {
	logger *zap.SugaredLogger
}

func NewSigningFailureReporter(logger *zap.SugaredLogger) *SigningFailureReporter {
	return &SigningFailureReporter{logger: logger}
}

func (r *SigningFailureReporter) ReportSigningFailure(err error) {
	r.logger.Warnw("failed to sign textures property", "error", err)
}
