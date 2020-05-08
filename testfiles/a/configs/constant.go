// THIS FILE IS A GENERATED CODE. DO NOT EDIT
package configs

import "strconv"

type BoolCriteria string

const (
	BoolCriteriaEmpty BoolCriteria = ""
	BoolCriteriaTrue  BoolCriteria = "true"
	BoolCriteriaFalse BoolCriteria = "false"
)

func (src BoolCriteria) Bool() bool {
	return src == BoolCriteriaTrue
}

type IntegerCriteria string

const (
	IntegerCriteriaEmpty IntegerCriteria = ""
)

func (str IntegerCriteria) Int() int {
	i32, err := strconv.Atoi(string(str))
	if err != nil {
		return -1
	}
	return i32
}

func (str IntegerCriteria) Int64() int64 {
	i64, err := strconv.ParseInt(string(str), 10, 64)
	if err != nil {
		return -1
	}
	return i64
}
