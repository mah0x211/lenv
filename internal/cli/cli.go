package cli

import (
	"fmt"
	"os"
	"runtime"
)

func Eprintf(format string, v ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(format, v...))
	os.Stderr.Write([]byte{'\n'})
}

var ExitCode int

func Fatalf(format string, v ...interface{}) {
	ExitCode = 1
	Eprintf(format, v...)
	runtime.Goexit()
}

func Errorf(format string, v ...interface{}) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("failed to runtime,Caller()")
	}
	fname := runtime.FuncForPC(pc).Name()

	return fmt.Errorf("[%s:%s:%d] "+format, append([]interface{}{file, fname, line}, v...)...)
}

func Print(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func Printf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func Println(v ...interface{}) {
	fmt.Println(v...)
}
