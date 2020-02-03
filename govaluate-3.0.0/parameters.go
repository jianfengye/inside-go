package govaluate

import (
	"errors"
)

/*
	Parameters is a collection of named parameters that can be used by an EvaluableExpression to retrieve parameters
	when an expression tries to use them.
	将Map封装成一个实现Parameters接口的MapParameters对象，主要就是为了获取参数方便。
*/
type Parameters interface {

	/*
		Get gets the parameter of the given name, or an error if the parameter is unavailable.
		Failure to find the given parameter should be indicated by returning an error.
		获取key为name的参数，如果name不存在，则返回nil和error
	*/
	Get(name string) (interface{}, error)
}

// MapParameters对象
type MapParameters map[string]interface{}

func (p MapParameters) Get(name string) (interface{}, error) {

	value, found := p[name]

	if !found {
		errorMessage := "No parameter '" + name + "' found."
		return nil, errors.New(errorMessage)
	}

	return value, nil
}
