package service

import (
	"context"
	"xo-packs/model"
	"xo-packs/repository"
)

type LoggingService interface {
	LogSignIn(context.Context, string, string) (*model.SignInLog, error)
	LogAgeAgreement(context.Context, string, string) (*model.AgeAgreementLog, error)
}

type LoggingServiceImpl struct {
	LoggingRepo repository.LoggingRepository
}

func NewLoggingService(loggingRepo repository.LoggingRepository) LoggingService {
	return &LoggingServiceImpl{LoggingRepo: loggingRepo}
}

func (s *LoggingServiceImpl) LogSignIn(c context.Context, uid string, ip string) (*model.SignInLog, error) {
	return s.LoggingRepo.LogSignIn(c, uid, ip)
}

func (s *LoggingServiceImpl) LogAgeAgreement(c context.Context, uid string, ip string) (*model.AgeAgreementLog, error) {
	return s.LoggingRepo.LogAgeAgreement(c, uid, ip)
}
