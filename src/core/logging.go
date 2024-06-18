package core

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AddLog(fields logrus.Fields, ctx *gin.Context, infoMsg string) error {
	logEntry, exists := ctx.Get("requestLogEntry")

	if exists {
		requestLogEntry, ok := logEntry.(*logrus.Entry)
		if ok {
			userLogEntry := requestLogEntry.WithFields(fields)
			userLogEntry.Info(infoMsg)
			return nil
		}
		return &ErrorResp{Message: "could not cast request log entry to *logrus.Entry "}
	}
	return &ErrorResp{Message: "request log entry does not exist"}
}
