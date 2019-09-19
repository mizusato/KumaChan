package object

func __DoSortedIntSlicesHaveContainingRelationship (A []int, B []int) bool {
	if len(A) == 0 {
		return true
	} else if len(A) > len(B) {
		return false
	}
	// 0 < len(A) <= len(B)
	var ok = true
	var i = 0
	for _, current := range A {
		// try to find current in B
		for i < len(B) && B[i] != current {
			i += 1
		}
		if i == len(B) {
			// not found
			ok = false
			break
		} else {
			// found
			i += 1
		}
	}
	return ok
}