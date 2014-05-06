// Package hiertree arranges a list of flat paths (and associated objects) into a hierarchical tree.
package hiertree

import (
	"fmt"
	"sort"
	"strings"
)

// Elem represents an object with a path.
type Elem interface {
	// HierPath is the object's path in the tree, with path components separated by slashes
	// (e.g., "a/b/c").
	HierPath() string
}

// Entry represents an entry in the resulting hierarchical tree. Elem is nil if the entry is a
// stub (i.e., no element exists at that path, but it does contain elements).
type Entry struct {
	// Parent is the full path of this entry's parent
	Parent string

	// Name is the name of this entry (without the parent path)
	Name string

	// Elem is the element that exists at this position in the tree, or nil if the entry is a stub
	Elem Elem

	// Leaf is true iff this entry is a leaf node (i.e., it has no children)
	Leaf bool
}

// List arranges elems into a flat list based on their hierarchical paths.
func List(elems []Elem) ([]Entry, error) {
	nodes, err := Tree(elems)
	if err != nil {
		return nil, err
	}
	return list(nodes, "")
}

func list(nodes []Node, parent string) ([]Entry, error) {
	var entries []Entry
	for _, n := range nodes {
		entries = append(entries, Entry{
			Parent: parent,
			Name:   n.Name,
			Elem:   n.Elem,
			Leaf:   len(n.Children) == 0,
		})
		var prefix string
		if parent == "" {
			prefix = n.Name
		} else {
			prefix = parent + "/" + n.Name
		}
		var children []Entry
		children, err := list(n.Children, prefix)
		if err != nil {
			return nil, err
		}
		entries = append(entries, children...)
	}
	return entries, nil
}

// Node represents a node in the resulting tree. Elem is nil if the entry is a stub (i.e., no
// element exists at this path, but it does contain elements).
type Node struct {
	// Name is the name of this node (without the parent path)
	Name string

	// Elem is the element that exists at this position in the tree, or nil if the entry is a stub
	Elem Elem

	// Children is the list of child nodes under this node
	Children []Node
}

// Tree arranges elems into a tree based on their hierarchical paths.
func Tree(elems []Elem) ([]Node, error) {
	nodes, _, err := tree(elems, "")
	return nodes, err
}

func tree(elems []Elem, prefix string) (roots []Node, size int, err error) {
	es := elemlist(elems)
	if prefix == "" { // only sort on first call
		sort.Sort(es)
	}
	var cur *Node
	var saveCur = func() {
		if cur != nil {
			if cur.Elem != nil {
				size++
			}
			roots = append(roots, *cur)
		}
		cur = nil
	}
	defer saveCur()
	for i := 0; i < len(es); i++ {
		e := es[i]
		path := e.HierPath()
		if !strings.HasPrefix(path, prefix) {
			return roots, size, nil
		}
		relpath := path[len(prefix):]
		root, rest := split(relpath)
		if root == "" && err == nil {
			return nil, 0, fmt.Errorf("invalid node path: %q", path)
		}
		if cur != nil && cur.Name == relpath && err == nil {
			return nil, 0, fmt.Errorf("duplicate node path: %q", path)
		}
		if cur == nil || cur.Name != root {
			saveCur()
			cur = &Node{Name: root}
		}
		if rest == "" {
			cur.Elem = e
		}
		var n int
		cur.Children, n, err = tree(elems[i:], prefix+root+"/")
		if err != nil {
			return nil, 0, err
		}
		size += n
		if n > 0 {
			i += n - 1
		}
	}
	return roots, size, nil
}

type elemlist []Elem

func (vs elemlist) Len() int           { return len(vs) }
func (vs elemlist) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs elemlist) Less(i, j int) bool { return vs[i].HierPath() < vs[j].HierPath() }

// split splits path immediately following the first slash. The returned values have the property
// that path = root+"/"+rest.
func split(path string) (root, rest string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// Inspect returns a list of path strings of the form "[parent/]path*", where the asterisk indicates
// that the entry is not a stub.
func Inspect(entries []Entry) (paths []string) {
	paths = make([]string, len(entries))
	for i, e := range entries {
		if e.Parent != "" {
			paths[i] += "[" + e.Parent + "/]"
		}
		paths[i] += e.Name
		if e.Elem != nil {
			paths[i] += "*"
		}
		if !e.Leaf {
			paths[i] += ">"
		}
	}
	return
}
