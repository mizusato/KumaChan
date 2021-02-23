package lang


type Resource struct {
	Kind  string
	MIME  string
	Data  [] byte
}

func CategorizeResources(all (map[string] Resource)) (map[string] map[string] Resource) {
	var result = make(map[string] map[string] Resource)
	for path, item := range all {
		var this_kind_map, exists = result[item.Kind]
		if !(exists) {
			this_kind_map = make(map[string] Resource)
			result[item.Kind] = this_kind_map
		}
		this_kind_map[path] = item
	}
	return result
}

