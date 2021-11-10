package kubernetes

import (
	"fmt"
	"log"
)

type BuiltinLogger interface {
	Debugf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Printf(format string, args ...interface{})
}

type KubernetesLogger struct {
	Logger interface{}
}

func (logger *KubernetesLogger) Debugf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes DEBUG] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *KubernetesLogger) Warningf(format string, args ...interface{}) {
	if logger.Logger != nil {

		switch v := logger.Logger.(type) {
		case *log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes WARNING] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}

func (logger *KubernetesLogger) Printf(format string, args ...interface{}) {
	if logger.Logger != nil {
		switch v := logger.Logger.(type) {
		case log.Logger:
			finalFormat := fmt.Sprintf("[Kubernetes INFO] %s", format)
			v.Printf(finalFormat, args...)
		}
	}
}
