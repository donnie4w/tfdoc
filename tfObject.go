package tfdoc

import (
	"container/list"
)

type TFObjects []*TFObject

//struct exception
type TFObject struct {
	name       string
	body       string
	dependency *list.List
	score      int32
}

func newTfObjects() *TFObjects {
	return new(TFObjects)
}

func (this *TFObjects) add(tfo *TFObject) {
	*this = append(*this, tfo)
}

func (this TFObjects) Len() int {
	return len(this)
}

func (this TFObjects) Less(i, j int) bool {
	return this[i].score > this[j].score
}

func (this TFObjects) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
