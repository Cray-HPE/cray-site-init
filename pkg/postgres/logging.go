package postgres

import (
	"fmt"
	"log"
)

type BuiltinLogger interface {
	Debugf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Printf(format string, args ...interface{})
}

type PostgresLogger struct {
	Logger interface{}
}

func (logger *PostgresLogger) Debugf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Postgres DEBUG] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *PostgresLogger) Warningf(format string, args ...interface{}) {
	if logger.Logger != nil {

		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Postgres WARNING] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *PostgresLogger) Printf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case log.Logger:
			finalFormat := fmt.Sprintf("[Postgres INFO] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}
