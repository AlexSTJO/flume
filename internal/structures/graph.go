package structures

type Graph struct {
  Nodes map[string]Task
  Adj map[string][]string
  InDeg map[string]int
} 


func Build(p *Pipeline) (*Dag, error) {
  if p == nil || len(p.Tasks) == 0 {
    return nil, fmt.Errorf("Cannot build empty pipeline")
  }

  g := &Graph{
    Nodes: make(map[string]Task, len(p.Tasks)),
    Adj: make(map[string][]string, len(p.Tasks)),
    InDeg: make(map[string]int, len(p.Tasks)),
  }

  for name,t := range(p.Tasks) {
    g.Nodes[name] = t 
    g.InDeg[name] = 0
    
    seen := make(map[string]struct{}, len(p.Tasks))
    for _, dependency := range(t["Dependencies"]) {
      if dependency == name {
        return nil, fmt.Errorf("Task can't depend on itself")
      }
      _, ok := g[dependency]
      if !ok {
        return nil, fmt.Errorf("Dependency is not an existing task")
      } 

      if _, dup := seen[dup]; dup{
        continue
      }
      seen[dependency] = struct{}{}
      g.InDeg[name]++
      g.Adj[dependency] = append(g.adj[dependency], name)
    }
  }

  return g, nil

}
