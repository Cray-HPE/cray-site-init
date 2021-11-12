package postgres

import (
	"fmt"
	"log"
)

type postgresLogger struct {
	Logger interface{}
}

func (logger *postgresLogger) debugf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Postgres DEBUG] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *postgresLogger) warningf(format string, args ...interface{}) {
	if logger.Logger != nil {

		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Postgres WARNING] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *postgresLogger) printf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case log.Logger:
			finalFormat := fmt.Sprintf("[Postgres INFO] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}
