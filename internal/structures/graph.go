package structures


import (
  "fmt"
  "strings"
)

type Graph struct {
  Nodes map[string]Task
  Adj map[string][]string
  InDeg map[string]int
} 


func BuildGraph(p *Pipeline) (*Graph, error) {
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
    for _, dependency := range(t.Dependencies) {
      if dependency == name {
        return nil, fmt.Errorf("Task can't depend on itself")
      }
      _, ok := g.Nodes[dependency]
      if !ok {
        return nil, fmt.Errorf("Dependency is not an existing task")
      } 

      if _, dup := seen[dependency]; dup{
        continue
      }
      seen[dependency] = struct{}{}
      g.InDeg[name]++
      g.Adj[dependency] = append(g.Adj[dependency], name)
    }
  }

  return g, nil

}


func (g *Graph) Levels() ([][]string, error) {
  if g == nil {
    return nil, fmt.Errorf("Graph is empty")
  }

  in := make(map[string]int, len(g.Nodes))
  for n,v := range(g.InDeg) {
    in[n] = v
  }


  levels := make([][]string, len(g.Nodes))
  curr := make([]string, 0, len(g.Nodes))

  for n,v := range(in) {
      if v == 0 {
        curr = append(curr, strings.TrimSpace(n))
      }
  }


  for len(curr) > 0 {
    levels = append(levels, curr)
    next := []string{}

    for _, n := range(curr) {
      if len(g.Adj[n]) == 0 {
        continue
      }
      
      for _, a := range(g.Adj[n]) {
        g.InDeg[a]--
        if (g.InDeg[a] == 0) {
          next = append(next, a)
        }
    } 
    }
    curr=next
  }

  return levels, nil
}
