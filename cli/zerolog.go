package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/fatih/color"
	"github.com/k0kubun/pp/v3"
)

var currentGoPackage string

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.CallerMarshalFunc = ZeroLogCallerMarshalFunc

	inf, ok := debug.ReadBuildInfo()
	if ok {
		currentGoPackage = inf.Main.Path
	}
}

type DefaultLoggerOpts struct {
	Level       zerolog.Level
	CommandName string
}

func ApplyDefaultLoggerContext(ctx context.Context, opts *DefaultLoggerOpts) context.Context {
	ctx = DefaultLogger(opts).WithContext(ctx)

	return ctx
}

func DefaultLogger(opts *DefaultLoggerOpts) *zerolog.Logger {

	consoleOutput := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro, NoColor: false}

	consoleOutput.FormatMessage = func(i interface{}) string {
		if i == nil {
			return "nil"
		}

		return color.New(color.FgHiWhite).Sprintf("%s", i.(string))
	}

	pretty := pp.New()

	pretty.SetColorScheme(pp.ColorScheme{})
	prettyerr := pp.New()
	prettyerr.SetExportedOnly(false)

	consoleOutput.FormatFieldName = func(i interface{}) string {
		return color.New(color.Faint).Sprintf("%s", i) + color.New(color.FgHiGreen).Sprint("=")
	}

	consoleOutput.FormatFieldValue = func(i interface{}) string {

		switch i := i.(type) {
		case error:
			return prettyerr.Sprint(i)
		case []byte:
			var g interface{}
			err := json.Unmarshal(i, &g)
			if err != nil {
				return pretty.Sprint(string(i))
			} else {
				return fmt.Sprintf("[unmarshaled json of byte array] %s", pretty.Sprint(g))
			}
		case string:
			return color.New(color.Bold).Sprint(i)
		}

		return pretty.Sprint(i)
	}

	consoleOutput.FormatLevel = func(i interface{}) string {
		switch i := i.(string); i {
		case zerolog.LevelDebugValue:
			return color.New(color.Bold, color.FgHiBlue).Sprint("debug")
		case zerolog.LevelInfoValue:
			return color.New(color.Bold, color.FgHiGreen).Sprint("info ")
		case zerolog.LevelWarnValue:
			return color.New(color.Bold, color.FgHiYellow).Sprint("warn ")
		case zerolog.LevelTraceValue:
			return color.New(color.Bold, color.FgHiCyan).Sprint("trace")
		case zerolog.LevelErrorValue:
			return color.New(color.Bold, color.FgHiRed).Sprint("error")
		case zerolog.LevelFatalValue:
			return color.New(color.Bold, color.FgHiRed).Sprint("fatal")
		default:
			return color.New(color.Bold, color.FgHiRed).Sprint(i)
		}
	}

	consoleOutput.FormatCaller = func(i interface{}) string {
		return fmt.Sprintf("%s", i) + color.New(color.FgHiGreen).Sprint(" >")
	}

	consoleOutput.FormatTimestamp = func(i interface{}) string {
		return color.New(color.Faint).Sprintf("%s", time.Now().UTC().Format("15:04:05.000000"))
	}

	consoleOutput.PartsOrder = []string{"level", "time", "caller", "message"}

	l := zerolog.New(consoleOutput).With().Caller().Timestamp()

	if opts.CommandName != "" {
		l = l.Str("cmd", opts.CommandName)
	}

	out := l.Logger().Level(opts.Level)

	return &out
}

func ZeroLogCallerMarshalFunc(pc uintptr, file string, line int) string {
	pkg, _ := GetPackageAndFuncFromFuncName(runtime.FuncForPC(pc).Name())

	return FormatCaller(pkg, file, line)
}

func GetPackageAndFuncFromFuncName(pc string) (pkg, function string) {
	funcName := pc
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}

	firstDot := strings.IndexByte(funcName[lastSlash:], '.') + lastSlash

	pkg = funcName[:firstDot]
	fname := funcName[firstDot+1:]

	if strings.Contains(pkg, ".(") {
		splt := strings.Split(pkg, ".(")
		pkg = splt[0]
		fname = "(" + splt[1] + "." + fname
	}

	pkg = strings.TrimPrefix(pkg, currentGoPackage+"/")

	return pkg, fname
}

func FormatCaller(pkg, path string, number int) string {
	// pkg = filepath.Base(pkg)
	path = color.New(color.Bold).Sprint(FileNameOfPath(path))
	num := color.New(color.FgHiRed, color.Bold).Sprintf("%d", number)
	sep := color.New(color.Faint).Sprint(":")

	return fmt.Sprintf("%s%s%s%s%s", pkg, sep, path, sep, num)
}

func FileNameOfPath(path string) string {
	tot := strings.Split(path, "/")
	if len(tot) > 1 {
		return tot[len(tot)-1]
	}

	return path
}
