/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package raft

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"
	"ztna-core/ztna/logtrace"

	"github.com/hashicorp/go-hclog"
	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"
)

func NewHcLogrusLogger() hclog.Logger {
	logtrace.LogWithFunctionName()
	logger := logrus.New()
	logger.SetFormatter(pfxlog.Logger().Entry.Logger.Formatter)

	return &hclogAdapter{
		entry: logrus.NewEntry(logger),
	}
}

type hclogAdapter struct {
	entry *logrus.Entry
	sync.Mutex
	name string
}

func (self *hclogAdapter) GetLevel() hclog.Level {
	logtrace.LogWithFunctionName()
	switch self.entry.Logger.Level {
	case logrus.TraceLevel:
		return hclog.Trace
	case logrus.DebugLevel:
		return hclog.Debug
	case logrus.InfoLevel:
		return hclog.Info
	case logrus.WarnLevel:
		return hclog.Warn
	case logrus.ErrorLevel:
		return hclog.Error
	case logrus.FatalLevel:
		return hclog.Error
	}
	return hclog.DefaultLevel
}

func (self *hclogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	switch level {
	case hclog.Trace:
		self.Trace(msg, args...)
	case hclog.Debug:
		self.Debug(msg, args...)
	case hclog.Info:
		self.Info(msg, args...)
	case hclog.Warn:
		self.Warn(msg, args...)
	case hclog.Error:
		self.Error(msg, args...)
	case hclog.Off:
	}
}

func (self *hclogAdapter) ImpliedArgs() []interface{} {
	logtrace.LogWithFunctionName()
	var fields []interface{}
	for k, v := range self.entry.Data {
		fields = append(fields, k)
		fields = append(fields, v)
	}
	return fields
}

func (self *hclogAdapter) Name() string {
	logtrace.LogWithFunctionName()
	return self.name
}

func (self *hclogAdapter) Trace(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	self.logToLogrus(logrus.TraceLevel, msg, args...)
}

func (self *hclogAdapter) Debug(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	self.logToLogrus(logrus.DebugLevel, msg, args...)
}

func (self *hclogAdapter) Info(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	self.logToLogrus(logrus.InfoLevel, msg, args...)
}

func (self *hclogAdapter) Warn(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	self.logToLogrus(logrus.WarnLevel, msg, args...)
}

func (self *hclogAdapter) Error(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	self.logToLogrus(logrus.ErrorLevel, msg, args...)
}

func (self *hclogAdapter) logToLogrus(level logrus.Level, msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	logger := self.entry
	if len(args) > 0 {
		logger = self.LoggerWith(args)
	}
	frame := self.getCaller()
	logger = logger.WithField("file", frame.File).WithField("func", frame.Function)
	logger.Log(level, self.name+msg)
}

func (self *hclogAdapter) IsTrace() bool {
	logtrace.LogWithFunctionName()
	return self.entry.Logger.IsLevelEnabled(logrus.TraceLevel)
}

func (self *hclogAdapter) IsDebug() bool {
	logtrace.LogWithFunctionName()
	return self.entry.Logger.IsLevelEnabled(logrus.DebugLevel)
}

func (self *hclogAdapter) IsInfo() bool {
	logtrace.LogWithFunctionName()
	return self.entry.Logger.IsLevelEnabled(logrus.InfoLevel)
}

func (self *hclogAdapter) IsWarn() bool {
	logtrace.LogWithFunctionName()
	return self.entry.Logger.IsLevelEnabled(logrus.WarnLevel)
}

func (self *hclogAdapter) IsError() bool {
	logtrace.LogWithFunctionName()
	return self.entry.Logger.IsLevelEnabled(logrus.ErrorLevel)
}

func (self *hclogAdapter) With(args ...interface{}) hclog.Logger {
	logtrace.LogWithFunctionName()
	return &hclogAdapter{
		entry: self.LoggerWith(args),
	}
}

func (self *hclogAdapter) LoggerWith(args []interface{}) *logrus.Entry {
	logtrace.LogWithFunctionName()
	l := self.entry
	ml := len(args)
	var key string
	for i := 0; i < ml-1; i += 2 {
		keyVal := args[i]
		if keyStr, ok := keyVal.(string); ok {
			key = keyStr
		} else {
			key = fmt.Sprintf("%v", keyVal)
		}
		val := args[i+1]
		if f, ok := val.(hclog.Format); ok {
			val = fmt.Sprintf(f[0].(string), f[1:])
		}
		l = l.WithField(key, val)
	}
	return l
}

func (self *hclogAdapter) Named(name string) hclog.Logger {
	logtrace.LogWithFunctionName()
	return self.ResetNamed(name + self.name)
}

func (self *hclogAdapter) ResetNamed(name string) hclog.Logger {
	logtrace.LogWithFunctionName()
	return &hclogAdapter{
		name:  name,
		entry: self.entry,
	}
}

func (self *hclogAdapter) SetLevel(hclog.Level) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *hclogAdapter) StandardLogger(*hclog.StandardLoggerOptions) *log.Logger {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *hclogAdapter) StandardWriter(*hclog.StandardLoggerOptions) io.Writer {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

var (
	// qualified package name, cached at first use
	localPackage string

	// Positions in the call stack when tracing to report the calling method
	minimumCallerDepth = 1

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

const (
	maximumCallerDepth      int = 25
	knownLocalPackageFrames int = 4
)

// getCaller retrieves the name of the first non-logrus calling function
// derived from logrus code
func (self *hclogAdapter) getCaller() *runtime.Frame {
	logtrace.LogWithFunctionName()
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				localPackage = self.getPackageName(funcName)
				// fmt.Printf("local package: %v\n", localPackage)
				break
			}
		}

		minimumCallerDepth = knownLocalPackageFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := self.getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != localPackage {
			//fmt.Printf("frame func: %v\n", f.Function)
			return &f //nolint:scopelint
		}
	}

	// fmt.Printf("frame func not found\n")

	// if we got here, we failed to find the caller's context
	return nil
}

// derived from logrus code
// getPackageName reduces a fully qualified function name to the package name
// There really ought to be a better way...
func (self *hclogAdapter) getPackageName(f string) string {
	logtrace.LogWithFunctionName()
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
