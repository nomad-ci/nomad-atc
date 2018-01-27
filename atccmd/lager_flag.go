package atccmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"code.cloudfoundry.org/lager"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

type LagerFlag struct {
	LogLevel string `long:"log-level" default:"info" choice:"debug" choice:"info" choice:"error" choice:"fatal" description:"Minimum level of logs to see."`
}

func (f LagerFlag) Logger(component string) (lager.Logger, *lager.ReconfigurableSink) {
	var minLagerLogLevel lager.LogLevel
	switch f.LogLevel {
	case LogLevelDebug:
		minLagerLogLevel = lager.DEBUG
	case LogLevelInfo:
		minLagerLogLevel = lager.INFO
	case LogLevelError:
		minLagerLogLevel = lager.ERROR
	case LogLevelFatal:
		minLagerLogLevel = lager.FATAL
	default:
		panic(fmt.Sprintf("unknown log level: %s", f.LogLevel))
	}

	logger := lager.NewLogger(component)

	sink := lager.NewReconfigurableSink(NewPrettySink(os.Stdout, lager.DEBUG), minLagerLogLevel)
	logger.RegisterSink(sink)

	return logger, sink
}

type prettySink struct {
	writer      io.Writer
	minLogLevel lager.LogLevel
	writeL      *sync.Mutex
}

func NewPrettySink(writer io.Writer, minLogLevel lager.LogLevel) lager.Sink {
	return &prettySink{
		writer:      writer,
		minLogLevel: minLogLevel,
		writeL:      new(sync.Mutex),
	}
}

func (sink *prettySink) Log(log lager.LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}

	sink.writeL.Lock()

	var keys []string

	for k, _ := range log.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var pairs []string

	for _, k := range keys {
		var vs string
		vd, err := json.Marshal(log.Data[k])
		if err != nil {
			vs = fmt.Sprintf("%v", log.Data[k])
		} else {
			vs = string(vd)
		}

		pairs = append(pairs, k+"="+vs)
	}

	if log.Error == nil {
		if len(pairs) == 0 {
			fmt.Fprintf(sink.writer, "+ %s %s %s\n",
				log.Timestamp, log.Source, log.Message)
		} else {
			fmt.Fprintf(sink.writer, "+ %s %s %s\n  %s\n",
				log.Timestamp, log.Source, log.Message, strings.Join(pairs, " "))
		}
	} else {
		fmt.Fprintf(sink.writer, "! %s %s %s '%s'\n  %s\n",
			log.Timestamp, log.Source, log.Message, log.Error.Error(), strings.Join(pairs, " "))

	}
	sink.writeL.Unlock()
}
