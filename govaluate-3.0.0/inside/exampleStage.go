package main

import (
	"fmt"
)


type Operator func(input1 interface{}, input2 interface{}) interface{}

func AddOperator(input1 interface{}, input2 interface{}) interface{} {
	a := input1.(int)
	b := input2.(int)
	return a+b
}

func SubOperator(input1 interface{}, input2 interface{}) interface{} {
	a := input1.(int)
	b := input2.(int)
	return a-b
}

func MultiOperator(input1 interface{}, input2 interface{}) interface{} {
	a := input1.(int)
	b := input2.(int)
	return a*b
}



func makeNumberOperator(number interface{}) Operator {
	return func(input1 interface{}, input2 interface{}) interface{} {
		return number
	}
}

func callStage(stage *Stage) interface{} {
	if (stage.left == nil && stage.right == nil ) {
		return stage.Op(nil, nil)
	}
	return stage.Op(callStage(stage.left), callStage(stage.right))
}


type Stage struct {
	Op Operator
	left *Stage
	right *Stage
}


func main() {
	// 构造 1*(2+3)
	Stage1 := Stage {
		Op : makeNumberOperator(1),
		left : nil,
		right : nil,
	}
	Stage2 := Stage {
		Op : makeNumberOperator(2),
		left : nil,
		right : nil,
	}
	Stage3 := Stage {
		Op : makeNumberOperator(3),
		left : nil,
		right : nil,
	}
	
	right_Stage := Stage {
		Op : AddOperator,
		left: &Stage2,
		right: &Stage3,
	}
	
	root_Stage := &Stage {
		Op : MultiOperator,
		left: &Stage1,
		right: &right_Stage,
	}
	
	ret := callStage(root_Stage)
	fmt.Println(ret)
}
