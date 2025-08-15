package structures

type Service interface {
  Name() string
  Parameters() []string
  Run(t Task, n string, ctx *Context) error
}


var Registry = map[string]Service{}

