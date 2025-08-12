package structures

type Service interface {
  Name() string
  Parameters() []string
  Run(t Task) error
}


var Registry = map[string]Service{}

