package util


type Trie (map[rune] Trie)

func NewTrie() Trie {
	return Trie(make(map[rune] Trie))
}

func (t Trie) Insert(k []rune) {
	if t == nil { panic("cannot insert an item into nil trie") }
	if len(k) == 0 { return }
	var char = k[0]
	if t[char] == nil {
		t[char] = NewTrie()
	}
	t[char].Insert(k[1:])
}

func (t Trie) Lookup(k []rune, result *[][]rune) {
	t.lookup(k, make([]rune, 0), result)
}

func (t Trie) lookup(k []rune, path []rune, result *[][]rune) {
	if t == nil {
		if len(k) == 0 {
			*result = append(*result, path)
		}
	} else {
		for char, branch := range t {
			if len(k) == 0 || char == k[0] {
				var branch_path = make([]rune, len(path)+1)
				copy(branch_path, path)
				branch_path[len(path)] = char
				if len(k) == 0 {
					branch.lookup([]rune{}, branch_path, result)
				} else {
					branch.lookup(k[1:], branch_path, result)
				}
			}
		}
	}
}
