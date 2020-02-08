package lib

func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}

func max(a int, b int) int {
	if a >= b { return a } else { return b }
}
