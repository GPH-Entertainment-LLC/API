package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		reqMethod := ctx.Request.Method
		reqUri := ctx.Request.RequestURI
		forwardedHeader := ctx.Request.Header["Forwarded"]
		reqUrl := ctx.Request.URL

		forwardedValues := []string{}
		if len(forwardedHeader) > 0 {
			forwardedValues = strings.Split(forwardedHeader[0], ";")
		}

		clientIP := ""
		if len(forwardedValues) > 0 {
			clientIP = forwardedValues[0]
			clientIP = strings.TrimPrefix(clientIP, "for=")
		}

		logger := log.New()
		logger.SetLevel(log.InfoLevel)

		entry := log.WithFields(
			log.Fields{
				"CLIENT_IP":     clientIP,
				"METHOD":        reqMethod,
				"URI":           reqUri,
				"REQUEST_START": startTime,
				"URL":           reqUrl,
			},
		)
		entry.Info("HTTP Request Client Log Entry")

		// setting the request context logEntry key
		ctx.Set("requestLogEntry", entry)
		ctx.Next()
	}
}

func FulfilledRequestLoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()

		reqMethod := ctx.Request.Method
		reqUri := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		reqUrl := ctx.Request.URL

		forwardedHeader := ctx.Request.Header["Forwarded"]
		forwardedValues := []string{}
		if len(forwardedHeader) > 0 {
			forwardedValues = strings.Split(forwardedHeader[0], ";")
		}

		clientIP := ""
		if len(forwardedValues) > 0 {
			clientIP = forwardedValues[0]
			clientIP = strings.TrimPrefix(clientIP, "for=")
		}

		// building the request log entry
		entry := log.WithFields(
			log.Fields{
				"CLIENT_IP":     clientIP,
				"METHOD":        reqMethod,
				"URI":           reqUri,
				"STATUS":        statusCode,
				"REQUEST_START": startTime,
				"REQUEST_STOP":  endTime,
				"REQUEST URL":   reqUrl,
			},
		)
		entry.Info("HTTP Fulfilled Request Log Entry")

		// setting the request context logEntry key
		ctx.Set("requestFinalLogEntry", entry)
		ctx.Next()
	}
}

func writeLogToFile(logEntry *log.Entry) {
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatal("Error getting current working directory:", err)
	}
	fmt.Println("======== Working Dir: ", wd)
	filename := "../request-logs/" + time.Now().UTC().Format("2006-01-02") + ".txt"
	filePath := filepath.Join(wd, filename)
	fmt.Println(filePath)

	// Open the file for writing in append mode
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Error("Error opening log file:", err)
		return
	}
	defer file.Close()

	jsonBytes, err := json.Marshal(logEntry.Data)
	if err != nil {
		logrus.Error("Error serializing json log")
		return
	}

	// Write the log entry to the file
	if _, err := file.WriteString(string(jsonBytes) + "\n"); err != nil {
		logrus.Error("Error writing log entry to file:", err)
	}
}
