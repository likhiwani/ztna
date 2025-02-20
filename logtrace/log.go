package logtrace

import (
    "github.com/sirupsen/logrus"
    "runtime"
)

func LogWithFunctionName() {
    pc, file, line, ok := runtime.Caller(1)
    if !ok {
        logrus.Println("unknown function")
        return
    }

    fn := runtime.FuncForPC(pc)
    if fn == nil {
        logrus.Println("unknown function")
        return
    }

    logrus.Printf("File: %s, Line: %d, Function: %s\n", file, line, fn.Name())
}