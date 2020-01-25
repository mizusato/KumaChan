package checker

import "kumachan/transformer/node"

type TypeCheckError interface {
	GetNodes()    [] node.Node
	GetMessage()  string
}
