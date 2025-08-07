package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	retalog "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// Logger 日志中间件
// todo 可考虑更换其他日志中间件
func Logger() gin.HandlerFunc {
	filePath := "log/log"
	//linkName := "latest_log.log"

	// 优化1：处理os.OpenFile的错误（如果日志文件打开失败，直接终止程序）
	scr, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		// 使用fmt.Fprintf输出到标准错误流，并退出程序
		fmt.Fprintf(os.Stderr, "无法打开日志文件: %v\n", err)
		os.Exit(1) // 非0状态码表示程序异常退出
	}

	logger := logrus.New()

	logger.Out = scr

	logger.SetLevel(logrus.DebugLevel)
	// 优化2：处理retalog.New的错误（日志轮转初始化失败时终止程序）
	logWriter, err := retalog.New(
		filePath+"%Y%m%d.log",
		retalog.WithMaxAge(7*24*time.Hour),     // 日志保留7天
		retalog.WithRotationTime(24*time.Hour), // 每天切分一个新日志文件
		//retalog.WithLinkName(linkName),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志轮转初始化失败: %v\n", err)
		os.Exit(1)
	}

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}
	// 优化3：处理lfshook.NewHook的潜在错误（虽然概率低，但仍需处理）
	Hook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	if Hook == nil {
		fmt.Fprintf(os.Stderr, "创建日志Hook失败\n")
		os.Exit(1)
	}

	logger.AddHook(Hook)

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		stopTime := time.Since(startTime).Milliseconds()
		spendTime := fmt.Sprintf("%d ms", stopTime)
		hostName, err := os.Hostname()
		if err != nil {
			hostName = "unknown"
		}
		statusCode := c.Writer.Status()
		clientIp := c.ClientIP()
		userAgent := c.Request.UserAgent()
		dataSize := c.Writer.Size()
		if dataSize < 0 {
			dataSize = 0
		}
		method := c.Request.Method
		path := c.Request.RequestURI

		entry := logger.WithFields(logrus.Fields{
			"HostName":  hostName,
			"status":    statusCode,
			"SpendTime": spendTime,
			"Ip":        clientIp,
			"Method":    method,
			"Path":      path,
			"DataSize":  dataSize,
			"Agent":     userAgent,
		})
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}
		if statusCode >= 500 {
			entry.Error()
		} else if statusCode >= 400 {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
