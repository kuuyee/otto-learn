package goapp

import (
	"github.com/kuuyee/otto-learn/app"
)

// Tuple 是app的元数据
var Tuples = app.TupleSlice([]app.Tuple{
	{"go", "aws", "simple"},
	{"go", "aws", "vpc-public-private"},
})
