package structures

type Context struct {
  Events map[string]map[string]string
}

func NewContext() *Context {
  return &Context {
    Events: make(map[string]map[string]string),
  }
}

func (c *Context) SetEventValues(key string, values map[string]string) {
  c.Events[key] = values
}


func (c *Context) GetEventValues(key string) map[string]string {
  return c.Events[key]
}
