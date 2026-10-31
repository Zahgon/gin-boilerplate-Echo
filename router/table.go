// Package router holds the table the service matches a request's path against, and reports what
// the framework the service was built on would have done with a path that does not match exactly.
//
// The matching and redirect behaviour reproduced here is that of github.com/gin-gonic/gin
// v1.11.0 and its tree.go (MIT License), which is itself derived from
// github.com/julienschmidt/httprouter (BSD-3-Clause). No code is taken from either; only the
// behaviour a client can observe is reproduced.
package router

import "strings"

type kind uint8

const (
	static kind = iota
	root
	param
	catchAll
)

type node struct {
	segment   string
	indices   string
	children  []*node
	wildChild bool
	kind      kind
	handler   bool
}

type skipped struct {
	path string
	node *node
}

// Table mirrors the route table so that the two questions the router answers
// before any handler runs can be asked again: is this path routed, and if it
// is not, should the client be sent to the same path with its trailing slash
// added or removed.
type Table struct {
	trees map[string]*node
}

// New builds a table from the registered route patterns per method, written
// with ":name" for a path parameter and "*" for a catch-all.
func New(routes map[string][]string) *Table {
	table := &Table{trees: make(map[string]*node, len(routes))}
	for method, paths := range routes {
		tree := new(node)
		for _, path := range paths {
			tree.insert(path)
		}
		table.trees[method] = tree
	}
	return table
}

// Lookup reports whether path is routed for method and, when it is not,
// whether a trailing slash redirect is recommended for it.
func (t *Table) Lookup(method, path string) (routed, redirect bool) {
	tree, ok := t.trees[method]
	if !ok {
		return false, false
	}
	return tree.lookup(path)
}

func (n *node) insert(path string) {
	if n.segment == "" && len(n.children) == 0 {
		n.insertChild(path)
		n.kind = root
		return
	}

	for {
		shared := commonPrefixLength(path, n.segment)
		if shared < len(n.segment) {
			n.children = []*node{{
				segment:   n.segment[shared:],
				indices:   n.indices,
				children:  n.children,
				wildChild: n.wildChild,
				kind:      static,
				handler:   n.handler,
			}}
			n.indices = n.segment[shared : shared+1]
			n.segment = n.segment[:shared]
			n.handler = false
			n.wildChild = false
		}

		if shared >= len(path) {
			n.handler = true
			return
		}

		path = path[shared:]
		next := path[0]

		if n.kind == param && next == '/' && len(n.children) == 1 {
			n = n.children[0]
			continue
		}

		if index := strings.IndexByte(n.indices, next); index >= 0 {
			n = n.children[index]
			continue
		}

		if next != ':' && next != '*' && n.kind != catchAll {
			child := new(node)
			n.indices += string(next)
			n.addChild(child)
			child.insertChild(path)
			return
		}

		if n.wildChild {
			n = n.children[len(n.children)-1]
			continue
		}

		n.insertChild(path)
		return
	}
}

func (n *node) insertChild(path string) {
	for {
		wildcard, at := findWildcard(path)
		if at < 0 {
			break
		}

		if wildcard[0] == ':' {
			if at > 0 {
				n.segment = path[:at]
				path = path[at:]
			}
			child := &node{segment: wildcard, kind: param}
			n.addChild(child)
			n.wildChild = true
			n = child
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child = new(node)
				n.addChild(child)
				n = child
				continue
			}
			n.handler = true
			return
		}

		at = strings.LastIndexByte(path[:at], '/')
		n.segment = path[:at]
		gate := &node{wildChild: true, kind: catchAll}
		n.children = []*node{gate}
		n.indices = "/"
		gate.children = []*node{{segment: path[at:], kind: catchAll, handler: true}}
		return
	}

	n.segment = path
	n.handler = true
}

func (n *node) addChild(child *node) {
	if n.wildChild && len(n.children) > 0 {
		wild := n.children[len(n.children)-1]
		n.children = append(n.children[:len(n.children)-1], child, wild)
		return
	}
	n.children = append(n.children, child)
}

func (n *node) clone() *node {
	copied := *n
	return &copied
}

func (n *node) lookup(path string) (routed, redirect bool) {
	var trail []skipped

	for {
		prefix := n.segment

		if len(path) > len(prefix) && strings.HasPrefix(path, prefix) {
			path = path[len(prefix):]

			if index := strings.IndexByte(n.indices, path[0]); index >= 0 {
				if n.wildChild {
					trail = append(trail, skipped{path: prefix + path, node: n.clone()})
				}
				n = n.children[index]
				continue
			}

			if !n.wildChild {
				if path != "/" {
					if resumed, rest, ok := pop(&trail, path); ok {
						n, path = resumed, rest
						continue
					}
				}
				return false, path == "/" && n.handler
			}

			n = n.children[len(n.children)-1]
			if n.kind == catchAll {
				return n.handler, false
			}

			end := 0
			for end < len(path) && path[end] != '/' {
				end++
			}
			if end < len(path) {
				if len(n.children) > 0 {
					path = path[end:]
					n = n.children[0]
					continue
				}
				return false, len(path) == end+1
			}
			if n.handler {
				return true, false
			}
			if len(n.children) == 1 {
				only := n.children[0]
				return false, (only.segment == "/" && only.handler) ||
					(only.segment == "" && only.indices == "/")
			}
			return false, false
		}

		if path == prefix {
			if !n.handler && path != "/" {
				if resumed, rest, ok := pop(&trail, path); ok {
					n, path = resumed, rest
					continue
				}
			}
			if n.handler {
				return true, false
			}
			if path == "/" && n.kind != root && (n.wildChild || n.kind == static) {
				return false, true
			}
			if index := strings.IndexByte(n.indices, '/'); index >= 0 {
				child := n.children[index]
				return false, (len(child.segment) == 1 && child.handler) ||
					(child.kind == catchAll && child.children[0].handler)
			}
			return false, false
		}

		redirect = path == "/" ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handler)
		if !redirect && path != "/" {
			if resumed, rest, ok := pop(&trail, path); ok {
				n, path = resumed, rest
				continue
			}
		}
		return false, redirect
	}
}

func pop(trail *[]skipped, path string) (*node, string, bool) {
	for length := len(*trail); length > 0; length-- {
		candidate := (*trail)[length-1]
		*trail = (*trail)[:length-1]
		if strings.HasSuffix(candidate.path, path) {
			return candidate.node, candidate.path, true
		}
	}
	return nil, "", false
}

func findWildcard(path string) (string, int) {
	for start := 0; start < len(path); start++ {
		if path[start] != ':' && path[start] != '*' {
			continue
		}
		if end := strings.IndexByte(path[start+1:], '/'); end >= 0 {
			return path[start : start+1+end], start
		}
		return path[start:], start
	}
	return "", -1
}

func commonPrefixLength(a, b string) int {
	limit := min(len(a), len(b))
	i := 0
	for i < limit && a[i] == b[i] {
		i++
	}
	return i
}
