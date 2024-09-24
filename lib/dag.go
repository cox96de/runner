package lib

import "github.com/cockroachdb/errors"

type Node interface {
	ID() string
	Depends() []string
}

type node[T Node] struct {
	n        T
	pre      []*node[T]
	deepPre  []*node[T]
	post     []*node[T]
	deepPost []*node[T]
}

// DAG uses topological sort to build a directed acyclic graph.
type DAG[T Node] struct {
	nodes map[string]*node[T]
}

// NewDAG creates a new DAG.
func NewDAG[T Node](nodes ...T) (*DAG[T], error) {
	nodeMap := make(map[string]*node[T], len(nodes))
	for _, n := range nodes {
		if _, ok := nodeMap[n.ID()]; ok {
			return nil, errors.Errorf("duplicate node id %s", n.ID())
		}
		nodeMap[n.ID()] = &node[T]{
			n: n,
		}
	}
	for _, node := range nodeMap {
		depends := node.n.Depends()
		for _, depend := range depends {
			if depend == node.n.ID() {
				return nil, errors.Errorf("detect self dependency of '%s'", depend) // TODO
			}
			dn, ok := nodeMap[depend]
			if !ok {
				return nil, errors.Errorf("the dependency '%s' of '%s' not found", depend, node.n.ID())
			}
			dn.post = append(dn.post, node)
			node.pre = append(node.pre, dn)
		}
	}
	orderedList := make([]*node[T], 0, len(nodeMap))
	preLeft := make(map[*node[T]]int)
	for _, n := range nodeMap {
		preLeft[n] = len(n.pre)
		if len(n.pre) == 0 {
			orderedList = append(orderedList, n)
		}
	}
	for i := 0; i < len(orderedList); i++ {
		for _, n := range orderedList[i].post {
			preLeft[n]--
			if preLeft[n] == 0 {
				orderedList = append(orderedList, n)
			}
		}
	}
	if len(orderedList) != len(nodeMap) {
		return nil, errors.Errorf("detect cycle")
	}
	matrix := make(map[*node[T]]map[*node[T]]bool)
	// Because cause of the orderedList, we can ensure that the deepPre of a node is always before the node itself.
	for _, n := range orderedList {
		matrix[n] = map[*node[T]]bool{}
		for _, directPre := range n.pre {
			matrix[n][directPre] = true
			for _, deepPre := range directPre.deepPre {
				matrix[n][deepPre] = true
			}
		}
		n.deepPre = make([]*node[T], 0, len(matrix[n]))
		for deepPre := range matrix[n] {
			n.deepPre = append(n.deepPre, deepPre)
		}
	}
	for _, n := range orderedList {
		for _, n2 := range orderedList {
			if !matrix[n2][n] {
				continue
			}
			n.deepPost = append(n.deepPost, n2)
		}
	}
	return &DAG[T]{
		nodes: nodeMap,
	}, nil
}

// DeepPre returns the deep preceding nodes (directly preceding and indirectly preceding) of the given node.
func (d *DAG[T]) DeepPre(node string) ([]T, error) {
	n, ok := d.nodes[node]
	if !ok {
		return nil, errors.Errorf("node not found")
	}
	res := make([]T, 0, len(n.deepPre))
	for _, n := range n.deepPre {
		res = append(res, n.n)
	}
	return res, nil
}

// DeepPost returns the deep posting nodes (directly posting and indirectly posting) of the given node.
func (d *DAG[T]) DeepPost(node string) ([]T, error) {
	n, ok := d.nodes[node]
	if !ok {
		return nil, errors.Errorf("node not found")
	}
	res := make([]T, 0, len(n.deepPost))
	for _, n := range n.deepPost {
		res = append(res, n.n)
	}
	return res, nil
}
