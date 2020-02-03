# 如何用golang解析计算表达式

本文解读开源库govaluate。

我们很多时候无法使用代码直接计算一个表达式，那么这个时候，我们就非常需要一个“表达式”来进行计算。这种场景是存在的，比如在数据收集的时候，我们已经收集了很多数据了，但是是否报警的规则我们希望让用户或者开发同学自己定义，这个时候，我们希望用户或者开发同学填写的是一个“表达式”，而非代码。这个时候，我们就需要能解析表达式并且计算的能力了。

我们这里就说说一个开源表达式解析库，[govaluate](https://github.com/Knetic/govaluate) (https://github.com/Knetic/govaluate) 。这个库我是在看casbin项目的时候接触到的，原本以为它很小，但是阅读进去才觉得这个库设计和写法是非常巧妙的。

我这里把 govaluate 这个库做了一份注释版（关键函数进行了注释），并且画了这个库的 UML 图，有兴趣

表达式可以支持哪些运算符代表了它的计算能力强弱。

计算运算符: + - * / ** %
字符连接: +
位运算符: >> << | & ^ ~
符号: -
转置: !
逻辑运算符: ||  &&
三元运算符: ? : 
是否为空: ??
比较运算符: > < >= <=
表达式运算符: =~ !~
优先级括号: ()

这些都是开源的表达式解析库 govaluate 可以支持的运算符。

# 基础使用

govaluate 的基础使用如下:

```go

expression, err := NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
parameters := make(map[string]interface{}, 8)
parameters["requests_made"] = 100
parameters["requests_succeeded"] = 80

result, err := expression.Evaluate(parameters)  // result == false
```

通过传递一个参数map：parameters，然后计算预先设计的表达式的最终值，得出了最后的计算值，最后的计算值类型可以是 bool 或者 float。

# 原理

如果我们手写这个表达式解析库，我们会怎么写呢？首先是解析的过程，将整个表达式拆解出来，然后将拆解后的表达式构建出一个表达式树，

