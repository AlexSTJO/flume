package services

import (
)


type Service interface {
  Name() string
  Parameters() []string
}


var Registry = map[string]Service{}

