package services

type ShellService struct {}


func (s ShellService) Name() string {
  return "string"
}


func (s ShellService) Parameters() []string {
  return []string{"command"}
} 

func init() {
  Registry["shell"] = ShellService{}
}

