package logger

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// Logger provides minimal structured logging without external runtime deps.
type Logger struct {
	env string
}

func New(env string) *Logger {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	return &Logger{env: env}
}

func (l *Logger) Info(msg string, kv ...interface{}) {
	l.log("info", msg, kv...)
}

func (l *Logger) Error(msg string, kv ...interface{}) {
	l.log("error", msg, kv...)
}

func (l *Logger) Debug(msg string, kv ...interface{}) {
	if strings.EqualFold(l.env, "development") {
		l.log("debug", msg, kv...)
	}
}

func (l *Logger) log(level string, msg string, kv ...interface{}) {
	entry := map[string]interface{}{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}

	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok || key == "" {
			continue
		}
		entry[key] = kv[i+1]
	}

	b, err := json.Marshal(entry)
	if err != nil {
		log.Printf(`{"level":"error","msg":"logger_marshal_failed","error":%q}`+"\n", err.Error())
		return
	}
	log.Println(string(b))
}
