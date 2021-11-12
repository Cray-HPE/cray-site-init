package kubernetes

import (
	"fmt"
	"log"
)

type kubernetesLogger struct {
	Logger interface{}
}

func (logger *kubernetesLogger) debugf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes DEBUG] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *kubernetesLogger) warningf(format string, args ...interface{}) {
	if logger.Logger != nil {

		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes WARNING] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *kubernetesLogger) printf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes INFO] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}
