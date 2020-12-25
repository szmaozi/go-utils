/*
 * @Descripttion:
 * @version:
 * @Author: moo
 * @Date: 2020-10-20 17:49:17
 * @LastEditors: moo
 * @LastEditTime: 2020-10-30 19:06:30
 */

package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AppLogger struct {
	l    *zap.Logger
	s    *zap.SugaredLogger
	name string
}

var logger AppLogger
var atomConsole = zap.NewAtomicLevel()
var atomFile = zap.NewAtomicLevel()

//格式化日期
func formatEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()))
}

/*
// ltype: 0-console 1-file
func getLevel(ltype uint) zapcore.Level {
	sLevel, fLevel := setting.GetLogLevel()
	if ltype == 0 {
		fmt.Printf("Console display level = [%s]\n", sLevel)
		return getZapLevel(sLevel)
	} else {
		fmt.Printf("Logfile record level = [%s]\n", fLevel)
		return getZapLevel(fLevel)
	}
}
*/

func getZapLevel(sLevel string) zapcore.Level {
	s := strings.ToLower(sLevel)
	switch s {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		fmt.Printf("logger level invaild[%s]! set default level-[INFO]\n", sLevel)
		return zapcore.InfoLevel
	}
}

// func getLevelName(level zapcore.Level) string {
// 	switch level {
// 	case zapcore.DebugLevel:
// 		return "debug"
// 	case zapcore.InfoLevel:
// 		return "info"
// 	case zapcore.WarnLevel:
// 		return "warn"
// 	case zapcore.ErrorLevel:
// 		return "error"
// 	case zapcore.DPanicLevel:
// 		return "dpanic"
// 	case zapcore.PanicLevel:
// 		return "panic"
// 	case zapcore.FatalLevel:
// 		return "fatal"
// 	default:
// 		fmt.Printf("logger level invaild[%d]! return default level-[INFO]\n", level)
// 		return "info"
// 	}
// }

func LogInit() {
	if logger.l != nil {
		return
	}
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		NameKey:    "logger",
		CallerKey:  "caller",
		MessageKey: "msg",
		//StacktraceKey: "stacktrace",
		LineEnding: zapcore.DefaultLineEnding,
		//EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeLevel: zapcore.CapitalColorLevelEncoder, //这里可以指定颜色
		// EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601 UTC 时间格式
		EncodeTime:     formatEncodeTime, // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		// EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:       "T",
		LevelKey:      "L",
		NameKey:       "N",
		CallerKey:     "C",
		MessageKey:    "M",
		StacktraceKey: "S",
		LineEnding:    zapcore.DefaultLineEnding,
		//EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeLevel:    zapcore.CapitalLevelEncoder, //这里可以指定颜色
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		// EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	/***
	// define our level-handling logic.
	// 仅打印Error级别以上的日志
	// sLevel, fLevel := setting.GetLogLevel()
	// fmt.Printf("Console display level = [%s], Logfile record level = [%s]\n", sLevel, fLevel)
	sLevelZap := getLevel(0)
	fLevelZap := getLevel(1)
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		// return lvl >= zapcore.ErrorLevel
		// return lvl >= getLevel(fLevel)	// 每次调用都会判断，性能低，但可动态调整
		return lvl >= fLevelZap
	})
	// 打印所有级别的日志
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		// return lvl >= getLevel(sLevel)
		return lvl >= sLevelZap
	})
	***/

	hook := lumberjack.Logger{
		Filename:   "./logs/logs",
		MaxSize:    1024, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	//topicErrors := zapcore.AddSync(ioutil.Discard)
	fileWriter := zapcore.AddSync(&hook)

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)

	// fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	// consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	SetDefaultLevel()
	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the four cores together.
	core := zapcore.NewTee(
		// 打印在topic中（伪造的case）
		//zapcore.NewCore(kafkaEncoder, topicErrors, highPriority),
		// 打印在控制台
		zapcore.NewCore(consoleEncoder, consoleDebugging, atomConsole),
		// 打印在文件中
		zapcore.NewCore(fileEncoder, fileWriter, atomFile),
	)
	// logger.l = zap.New(core, zap.AddCaller())
	logger.l = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger.s = logger.l.Sugar()
	logger.name = "main"
	fmt.Println("logging init success")
}

func SetLevel(levelConsole string, levelFile string) {
	atomConsole.SetLevel(getZapLevel(levelConsole))
	atomFile.SetLevel(getZapLevel(levelFile))
}

func GetLevel() (string, string) {
	s1 := atomConsole.Level().String()
	s2 := atomFile.Level().String()
	return s1, s2
}

func SetDefaultLevel() {
	atomConsole.SetLevel(zap.DebugLevel)
	atomFile.SetLevel(zap.InfoLevel)
	fmt.Println("logging set default level: console-[debug], file-[info]")
}

func MustGetLogger(names ...string) *AppLogger {
	if logger.l == nil {
		LogInit()
	}
	if len(names) > 0 {
		l := logger.l.Named(names[0])
		return &AppLogger{l: l, s: l.Sugar(), name: names[0]}
	} else {
		return &logger
	}
}

func Exit() {
	logger.Debug("logging exit")
	if logger.l == nil {
		return
	}
	logger.l.Sync()
	logger.l = nil
}

func (l *AppLogger) GetName() string {
	if l == nil {
		return ""
	}
	return l.name
}

func (l *AppLogger) GetZapLogger() *zap.Logger {
	return l.l
}

func (l *AppLogger) Sync() {
	l.s.Sync()
	l.l.Sync()
	return
}

func ZapLog(logger *zap.Logger, level string, msg string) {
	switch strings.ToLower(level) {
	case "debug":
		logger.Debug(msg)
	case "info":
		logger.Info(msg)
	case "warning":
		logger.Warn(msg)
	case "error":
		logger.Error(msg)
	case "panic", "fatal", "dpanic", "dfatal":
		logger.DPanic(msg)
	default:
		logger.Debug(msg)
	}
}

// func formatArgs(args []interface{}) string { return strings.TrimSuffix(fmt.Sprintln(args...), "\n") }

func (l *AppLogger) Debug(args ...interface{})                   { l.s.Debug(args...) }
func (l *AppLogger) Debugf(template string, args ...interface{}) { l.s.Debugf(template, args...) }
func (l *AppLogger) Info(args ...interface{})                    { l.s.Info(args...) }
func (l *AppLogger) Infof(template string, args ...interface{})  { l.s.Infof(template, args...) }
func (l *AppLogger) Warn(args ...interface{})                    { l.s.Warn(args...) }
func (l *AppLogger) Warnf(template string, args ...interface{})  { l.s.Warnf(template, args...) }
func (l *AppLogger) Error(args ...interface{})                   { l.s.Error(args...) }
func (l *AppLogger) Errorf(template string, args ...interface{}) { l.s.Errorf(template, args...) }
func (l *AppLogger) DPanic(args ...interface{})                  { l.s.DPanic(args...); l.s.Sync() }
func (l *AppLogger) DPanicf(template string, args ...interface{}) {
	l.s.DPanicf(template, args...)
	l.s.Sync()
}
func (l *AppLogger) Panic(args ...interface{}) { l.s.Panic(args...); l.s.Sync() }
func (l *AppLogger) Panicf(template string, args ...interface{}) {
	l.s.Panicf(template, args...)
	l.s.Sync()
}
func (l *AppLogger) Fatal(args ...interface{}) { l.s.Fatal(args...); l.s.Sync() }
func (l *AppLogger) Fatalf(template string, args ...interface{}) {
	l.s.Fatalf(template, args...)
	l.s.Sync()
}
