package logger

import (
	"path/filepath"
	"sync/atomic"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var emptyStr = ""

var log *Logger

func SetDefaultLog(l *Logger) {
	//l.getLog().WithOptions(zap.AddCallerSkip(-1))
	log = l
}

func New(name, path string, lvl zapcore.Level) (*Logger, error) {
	err := CheckAndCreate(path)
	if err != nil {
		return nil, err
	}

	return &Logger{
		tomorrow: unsafe.Pointer(&emptyStr),
		name:     name,
		path:     path,
		lvl:      lvl,
	}, nil
}

func newLogger(path string, lvl zapcore.Level) (*zap.Logger, error) {
	ec := zap.NewDevelopmentEncoderConfig()
	//ec.EncodeCaller = zapcore.FullCallerEncoder
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    ec,
		OutputPaths:      []string{"stderr", path},
		ErrorOutputPaths: []string{"stderr"},
	}
	log, err := cfg.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return log, nil
}

type Logger struct {
	log      unsafe.Pointer
	sugar    unsafe.Pointer
	tomorrow unsafe.Pointer
	name     string
	path     string
	t        time.Time
	lvl      zapcore.Level
}

func (this_ *Logger) SetCheckTomorrowTime(t time.Time) {
	this_.t = t
}

func (this_ *Logger) checkTomorrow(t time.Time) {
	if t.IsZero() {
		t = time.Now()
	}

	tomorrowStr := t.Format("2006-1-2")
	if tomorrowStr != *(*string)(atomic.LoadPointer(&this_.tomorrow)) {
		if atomic.CompareAndSwapPointer(&this_.tomorrow, this_.tomorrow, unsafe.Pointer(&tomorrowStr)) {
			pathFile := filepath.Join(this_.path, tomorrowStr+" "+this_.name+".log")
			log, err := newLogger(pathFile, this_.lvl)
			if err != nil {
				this_.Error("create new tomorrow logger failed", zap.Error(err))
				return
			}
			l := (*zap.Logger)(atomic.LoadPointer(&this_.log))
			if l != nil {
				_ = l.Sync()
			}

			atomic.StorePointer(&this_.log, unsafe.Pointer(log))
			atomic.StorePointer(&this_.sugar, unsafe.Pointer(log.Sugar()))
		}
	}
}
func (this_ *Logger) getLog() *zap.Logger {
	this_.checkTomorrow(this_.t)
	return (*zap.Logger)(atomic.LoadPointer(&this_.log))
}

func (this_ *Logger) getSugar() *zap.SugaredLogger {
	this_.checkTomorrow(this_.t)
	return (*zap.SugaredLogger)(atomic.LoadPointer(&this_.sugar))
}

func (this_ *Logger) Debug(msg string, fields ...zap.Field) {
	this_.getLog().Debug(msg, fields...)
}

func (this_ *Logger) Info(msg string, fields ...zap.Field) {
	this_.getLog().Info(msg, fields...)
}
func (this_ *Logger) Warn(msg string, fields ...zap.Field) {
	this_.getLog().Warn(msg, fields...)
}
func (this_ *Logger) Error(msg string, fields ...zap.Field) {
	this_.getLog().Error(msg, fields...)
}
func (this_ *Logger) DPanic(msg string, fields ...zap.Field) {
	this_.getLog().DPanic(msg, fields...)
}
func (this_ *Logger) Panic(msg string, fields ...zap.Field) {
	this_.getLog().Panic(msg, fields...)
}
func (this_ *Logger) Fatal(msg string, fields ...zap.Field) {
	this_.getLog().Fatal(msg, fields...)
}

func (this_ *Logger) DebugFormat(format string, args ...interface{}) {
	this_.getSugar().Debugf(format, args...)
}
func (this_ *Logger) InfoFormat(format string, args ...interface{}) {
	this_.getSugar().Infof(format, args...)
}
func (this_ *Logger) WarnFormat(format string, args ...interface{}) {
	this_.getSugar().Warnf(format, args...)
}
func (this_ *Logger) ErrorFormat(format string, args ...interface{}) {
	this_.getSugar().Errorf(format, args...)
}
func (this_ *Logger) DPanicFormat(format string, args ...interface{}) {
	this_.getSugar().DPanicf(format, args...)
}
func (this_ *Logger) PanicFormat(format string, args ...interface{}) {
	this_.getSugar().Panicf(format, args...)
}
func (this_ *Logger) FatalFormat(format string, args ...interface{}) {
	this_.getSugar().Fatalf(format, args...)
}

/* 暂时不开放出来用
func (this_ *Logger) DebugAny(args ...interface{}) {
	this_.getSugar().Debug(args...)
}
func (this_ *Logger) InfoAny(args ...interface{}) {
	this_.getSugar().Info(args...)
}
func (this_ *Logger) WarnAny(args ...interface{}) {
	this_.getSugar().Warn(args...)
}
func (this_ *Logger) ErrorAny(args ...interface{}) {
	this_.getSugar().Error(args...)
}
func (this_ *Logger) DPanicAny(args ...interface{}) {
	this_.getSugar().DPanic(args...)
}
func (this_ *Logger) PanicAny(args ...interface{}) {
	this_.getSugar().Panic(args...)
}
func (this_ *Logger) FatalAny(args ...interface{}) {
	this_.getSugar().Fatal(args...)
}
*/

func Debug(msg string, fields ...zap.Field) {
	log.getLog().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	log.getLog().Info(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	log.getLog().Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	log.getLog().Error(msg, fields...)
}
func DPanic(msg string, fields ...zap.Field) {
	log.getLog().DPanic(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {
	log.getLog().Panic(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	log.getLog().Fatal(msg, fields...)
}

func DebugFormat(format string, args ...interface{}) {
	log.getSugar().Debugf(format, args...)
}
func InfoFormat(format string, args ...interface{}) {
	log.getSugar().Infof(format, args...)
}
func WarnFormat(format string, args ...interface{}) {
	log.getSugar().Warnf(format, args...)
}
func ErrorFormat(format string, args ...interface{}) {
	log.getSugar().Errorf(format, args...)
}
func DPanicFormat(format string, args ...interface{}) {
	log.getSugar().DPanicf(format, args...)
}
func PanicFormat(format string, args ...interface{}) {
	log.getSugar().Panicf(format, args...)
}
func FatalFormat(format string, args ...interface{}) {
	log.getSugar().Fatalf(format, args...)
}

/*
func DebugAny(args ...interface{}) {
	log.getSugar().Debug(args...)
}
func InfoAny(args ...interface{}) {
	log.getSugar().Info(args...)
}
func WarnAny(args ...interface{}) {
	log.getSugar().Warn(args...)
}
func ErrorAny(args ...interface{}) {
	log.getSugar().Error(args...)
}
func DPanicAny(args ...interface{}) {
	log.getSugar().DPanic(args...)
}
func PanicAny(args ...interface{}) {
	log.getSugar().Panic(args...)
}
func FatalAny(args ...interface{}) {
	log.getSugar().Fatal(args...)
}
*/
