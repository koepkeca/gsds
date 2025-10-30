// Package trie implements a thread safe trie in go
package trie

import (
	"context"
	"fmt"
	"sort"
)

// lexicalKeys store the keys of a resulting depth search in lexical order.
// this is useful for obtaining a lexically sorted list from a given prefix
// search term. It implements the sort interface for a slice of runes.
type lexicalKeys []rune

func (lk lexicalKeys) Len() int           { return len(lk) }
func (lk lexicalKeys) Swap(i, j int)      { lk[i], lk[j] = lk[j], lk[i] }
func (lk lexicalKeys) Less(i, j int) bool { return lk[i] < lk[j] }

// T is the structure that contains the channel used to
// communicate with the trie.
type T struct {
	op      chan (func(*trie))
	cancel  context.CancelFunc
	Replace bool
}

// Exists checks if a key exists in the trie (has data associated with it)
func (t *T) Exists(k string) (exists bool, e error) {
	if k == "" {
		e = fmt.Errorf("Exists requires a string to search for")
		return
	}
	rch := make(chan bool)
	t.op <- func(st *trie) {
		curr := st.root
		for _, char := range k {
			next, ok := curr.children[char]
			if !ok {
				rch <- false
				return
			}
			curr = next
		}
		rch <- len(curr.data) > 0
		return
	}
	exists = <-rch
	return
}

// Insert will insert the data into the trie
// if Replace is true, the entire element will be replaced by the contents
// of v. Otherwise, the contents of v will be appended to the key.
func (t *T) Insert(k string, v []interface{}) (e error) {
	if k == "" {
		e = fmt.Errorf("Insert may not have empty key.")
		return
	}
	if len(v) == 0 {
		e = fmt.Errorf("Insert may not have empty data.")
		return
	}
	t.op <- func(st *trie) {
		curr := st.root
		for _, char := range k {
			if next, ok := curr.children[char]; !ok {
				nc := newNode()
				curr.children[char] = nc
				curr = nc
			} else {
				curr = next
			}
		}
		if t.Replace {
			curr.data = v
			return
		}
		curr.data = append(curr.data, v...)
		return
	}
	return
}

// Get retreives the data (if any) at location k
func (t *T) Get(k string) (v []interface{}, e error) {
	if k == "" {
		e = fmt.Errorf("Get requires a string to search for")
		return
	}
	ich := make(chan []interface{})
	t.op <- func(st *trie) {
		curr := st.root
		for _, char := range k {
			next, ok := curr.children[char]
			if !ok {
				ich <- nil
				return
			}
			curr = next
		}
		ich <- curr.data
		return
	}
	v = <-ich
	return
}

// Delete removes the key k from the trie and cleans up any unnecessary nodes
func (t *T) Delete(k string) (deleted bool, e error) {
	if k == "" {
		e = fmt.Errorf("Delete may not have empty key.")
		return
	}
	rch := make(chan bool)
	t.op <- func(st *trie) {
		runes := []rune(k)
		if len(runes) == 0 {
			rch <- false
			return
		}
		path := make([]*node, 0, len(runes)+1)
		curr := st.root
		path = append(path, curr)
		for _, char := range runes {
			next, ok := curr.children[char]
			if !ok {
				rch <- false
				return
			}
			curr = next
			path = append(path, curr)
		}
		if len(curr.data) == 0 {
			rch <- false
			return
		}
		curr.data = nil
		// Clean up unnecessary nodes from the end of the path backwards
		for i := len(path) - 1; i > 0; i-- {
			node := path[i]
			char := runes[i-1]
			parent := path[i-1]
			// If node has no data and no children, remove it
			if len(node.data) == 0 && len(node.children) == 0 {
				delete(parent.children, char)
				continue
			}
			break
		}
		rch <- true
		return
	}
	deleted = <-rch
	return
}

// SearchKeys returns all keys that start with the given prefix
// Returns keys in lexical order. Empty prefix returns all keys in the trie.
// Keys may be deleted between Search() and Get() calls.
// Callers should handle nil/missing results from Get().
func (t *T) Search(prefix string) []string {
	rch := make(chan []string)
	t.op <- func(st *trie) {
		curr := st.root
		for _, char := range prefix {
			next, ok := curr.children[char]
			if !ok {
				rch <- nil
				return
			}
			curr = next
		}
		keys := make([]string, 0, 64)
		curr.collectKeysBelow(prefix, &keys)
		rch <- keys
		return
	}
	return <-rch
}

// collectKeysBelow recursively collects all keys with data below this node
func (n *node) collectKeysBelow(currentKey string, keys *[]string) {
	if len(n.data) > 0 {
		*keys = append(*keys, currentKey)
	}
	if len(n.children) == 0 {
		return
	}

	tmpKeys := make([]rune, 0, len(n.children))
	for k := range n.children {
		tmpKeys = append(tmpKeys, k)
	}
	sort.Sort(lexicalKeys(tmpKeys))

	for _, key := range tmpKeys {
		n.children[key].collectKeysBelow(currentKey+string(key), keys)
	}
}

// trie contains the locally available trie
type trie struct {
	root *node
}

// node is the structure that contains the node data
// for the trie
type node struct {
	data     []interface{}
	children map[rune]*node
}

// newNode is the method for creating a new node for the trie
func newNode() (n *node) {
	n = &node{}
	n.children = make(map[rune]*node)
	return
}

// loop is the method that runs the goroutine for the data structure
func (t *T) loop(c context.Context, r chan struct{}) {
	core := &trie{}
	core.root = newNode()
	ctx, cancel := context.WithCancel(context.Background())
	if c != nil {
		ctx, cancel = context.WithCancel(c)
	}
	t.cancel = cancel
	close(r)
	for {
		select {
		case op := <-t.op:
			op(core)
		case <-ctx.Done():
			cancel()
			close(t.op)
			return
		}
	}
}

// Close stops the running go routine
func (t *T) Close() {
	t.cancel()
	return
}

// New creates a new trie
func New() (t *T) {
	t = NewWithContext(nil)
	return
}

// New creates a new trie with a context
func NewWithContext(ctx context.Context) (t *T) {
	t = &T{op: make(chan func(*trie))}
	ready := make(chan struct{})
	go t.loop(ctx, ready)
	<-ready
	return
}
