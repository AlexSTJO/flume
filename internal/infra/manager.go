package infra

import (
	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"

	"fmt"
)

type Service interface {
	Name() string
	Call(d structures.Deployment, r *structures.RunInfo, l *logging.Config) (map[string]string, error)
}

var registry = map[string]Service{}

func Deploy(i map[string]structures.Deployment, r *structures.RunInfo, l *logging.Config) (*map[string]map[string]string, error) {
	svc_outputs := make(map[string]map[string]string)
	for _, deployment := range i {
		svc, ok := registry[deployment.Service]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", deployment.Service)
		}
		outputs, err := svc.Call(deployment, r, l)
		if err != nil {
			return nil, err
		}
		svc_outputs[svc.Name()] = outputs
	}
	return &svc_outputs, nil
}
