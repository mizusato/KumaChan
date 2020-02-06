package vm

func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}
