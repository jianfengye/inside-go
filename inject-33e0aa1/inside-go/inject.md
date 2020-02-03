# golang实现依赖注入

依赖注入是软件工程中经常使用到的一种技术，它提供了一种控制反转的机制，把控制权利交给了调用方。调用方来决定使用哪些参数，哪些对象来进行具体的业务逻辑。

它有几个好处：
1 它让调用方更灵活。
2 大量减少定义类型的代码量
3 增加代码的可用性，因为调用方只需要关注它需要的参数，不需要顾及它不需要的参数了。

# 什么是依赖注入

依赖注入使用最多的应该是java中的spring框架了。依赖注入在使用的时候希望调用函数的参数是不固定的。
```go
function Action(a TypeA, b TypeB)
```

就是说，这个Action在实际调用的时候，可以任意加参数，每次加一个参数类型，都有一个容器可以给这个Action调用函数传递对应的参数对象提供使用。

# inject

Golang中也有项目是使用依赖注入实现的，martini就是一个依靠依赖注入实现的web框架，它的作者开源的https://github.com/codegangsta/inject 项目也就很值得我们学习。

这个inject项目很小，实际代码就一个文件，很容易阅读。

```go
// Injector代表依赖注入的容器需要实现的接口
type Injector interface {
	Applicator // 这个接口用来灌入到一个结构体
	Invoker    // 这个接口用来实际调用的，所以可以实现非反射的实际调用
	TypeMapper // 这个接口是真正的容器
	// SetParent sets the parent of the injector. If the injector cannot find a
	// dependency in its Type map it will check its parent before returning an
	// error.
	SetParent(Injector) // 表示这个结构是递归的
}
```

这个Injector使用三个接口进行组合，每个接口有各自不同的用处。

TypeMapper是依赖注入最核心的容器部分，注入类型和获取类型都是这个接口承载的。
Invoker和Applicator都是注入部分，Invoker将TypeMapper容器中的数据注入到调用函数中。而Applicator将容器中的数据注入到实体对象中。
最后我们还将Injector容器设计为有层级的，在我们获取容器数据的时候，会先从当前容器找，找不到再去父级别容器中找。

这几个接口中的TypeMapper又值得看一下：
```go
// TypeMapper represents an interface for mapping interface{} values based on type.
// TypeMapper是用来作为依赖注入容器的,设置的三种方法都是链式的
type TypeMapper interface {
	// Maps the interface{} value based on its immediate type from reflect.TypeOf.
	// 直接设置一个对象，TypeOf是key，value是这个对象
	Map(interface{}) TypeMapper
	// Maps the interface{} value based on the pointer of an Interface provided.
	// This is really only useful for mapping a value as an interface, as interfaces
	// cannot at this time be referenced directly without a pointer.
	// 将一个对象注入到一个接口中，TypeOf是接口，value是对象
	MapTo(interface{}, interface{}) TypeMapper
	// Provides a possibility to directly insert a mapping based on type and value.
	// This makes it possible to directly map type arguments not possible to instantiate
	// with reflect like unidirectional channels.
	// 直接手动设置key和value
	Set(reflect.Type, reflect.Value) TypeMapper
	// Returns the Value that is mapped to the current type. Returns a zeroed Value if
	// the Type has not been mapped.
	// 从容器中获取某个类型的注入对象
	Get(reflect.Type) reflect.Value
}
```
这里的Map是将数据注入，即将数据类型和数据值进行映射存储在容器中。MapTo是将数据接口和数据值进行映射存储在容器中。Set就是手动将数据类型活着数据接口和数据值存储在容器中。Get则和Set相反。

我们可以看下inject文件中实现了这个接口的对象：injector

```go
// 实际的注入容器，它实现了Injector的所有接口
type injector struct {
	// 这个就是容器最核心的map
	values map[reflect.Type]reflect.Value
	// 这里设置了一个parent，所以这个Inject是可以嵌套的
	parent Injector
}
```

其中的这个map[reflect.Type]reflect.Value就是最核心的。那么这里就需要注意到了，这个inject实际上是一个基础的map，而不是线程安全的map。所以如果在并发场景下，不应该在并发请求中进行动态注入或者改变容器元素。否则很有可能出现各种线程安全问题。

我们可以看看Map，Set等函数做的事情就是设置这个Map
```go
	i.values[reflect.TypeOf(val)] = reflect.ValueOf(val)
```

下一个重要的函数就Invoke。

这个Invoke做的事情我们也能很容易想清，根据它本身里面的函数参数类型，一个个去容器中拿对应值。
```go
// 真实的调用某个函数f，这里的f默认是function
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)

	var in = make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val := inj.Get(argType)
		if !val.IsValid() {
			return nil, fmt.Errorf("Value not found for type %v", argType)
		}

		in[i] = val
	}

	return reflect.ValueOf(f).Call(in), nil
}
```
注：inject相关的中文注释代码解读在项目：https://github.com/jianfengye/inside-go 中。

# go-macaron/inject

无闻在matini基础上又封装了一层inject。它使用的方法是直接保留CopyRight的通知，将https://github.com/codegangsta/inject 这个类做了一些修改。

我看了下这些修改，主要是增加了一个FastInvoker

```go
// FastInvoker represents an interface in order to avoid the calling function via reflection.
//
// example:
//	type handlerFuncHandler func(http.ResponseWriter, *http.Request) error
//	func (f handlerFuncHandler)Invoke([]interface{}) ([]reflect.Value, error){
//		ret := f(p[0].(http.ResponseWriter), p[1].(*http.Request))
//		return []reflect.Value{reflect.ValueOf(ret)}, nil
//	}
//
//	type funcHandler func(int, string)
//	func (f funcHandler)Invoke([]interface{}) ([]reflect.Value, error){
//		f(p[0].(int), p[1].(string))
//		return nil, nil
//	}
type FastInvoker interface {
	// Invoke attempts to call the ordinary functions. If f is a function
	// with the appropriate signature, f.Invoke([]interface{}) is a Call that calls f.
	// Returns a slice of reflect.Value representing the returned values of the function.
	// Returns an error if the injection fails.
	Invoke([]interface{}) ([]reflect.Value, error)
}
```

并且在Invoke调用的地方增加了一个分支，如果这个调用函数是自带有Invoke方法的，那么就用一种不用反射的方式。

```go
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)
	switch v := f.(type) {
	case FastInvoker:
		return inj.fastInvoke(v, t, t.NumIn())
	default:
		return inj.callInvoke(f, t, t.NumIn())
	}
}
```

我觉得这个fastInvoke是神来之笔啊。我们使用Golang的inject最害怕的就是性能问题。这里的Invoke频繁使用了反射，所以会导致Invoke的性能不会很高。但是我们有了fastInvoke替换方案，当需要追求性能的时候，我们就可以使用fastInvoke的方法进行替换。

# 示例

所以我下面的这个示例是最好的理解inject的例子：

```go
package main

import "gopkg.in/macaron.v1"

import "github.com/go-macaron/inject"

import "fmt"

import "reflect"

type A struct {
	Name string
}

type B struct {
	Name string
}

func (b *B) GetName() string {
	return b.Name
}

type I interface {
	GetName() string
}

type C struct {
	AStruct A `inject`
	BStruct B `inject`
}

type MyFastInvoker func(arg1 A, arg2 I, arg3 string)

func (invoker MyFastInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	if a, ok := args[0].(A); ok {
		fmt.Println(a.Name)
	}

	if b, ok := args[1].(I); ok {
		fmt.Println(b.GetName())
	}
	if c, ok := args[2].(string); ok {
		fmt.Println(c)
	}
	return nil, nil
}

type Invoker2 struct {
	inject.Injector
}

func main() {
	InjectDemo()

	a := &A{Name: "inject name"}
	m := macaron.Classic()
	m.Map(a)
	m.Get("/", func(a *A) string {
		return "Hello world!" + a.Name
	})
	m.Run()
}

func InjectDemo() {
	a := A{Name: "a name"}
	inject1 := inject.New()
	inject1.Map(a)
	inject1.MapTo(&B{Name: "b name"}, (*I)(nil))
	inject1.Set(reflect.TypeOf("string"), reflect.ValueOf("c name"))
	inject1.Invoke(func(arg1 A, arg2 I, arg3 string) {
		fmt.Println(arg1.Name)
		fmt.Println(arg2.GetName())
		fmt.Println(arg3)
	})

	c := C{}
	inject1.Apply(&c)
	fmt.Println(c.AStruct.Name)

	inject2 := inject.New()
	inject2.Map(a)
	inject2.MapTo(&B{Name: "b name"}, (*I)(nil))
	inject2.Set(reflect.TypeOf("string"), reflect.ValueOf("c name"))
	inject2.Invoke(MyFastInvoker(nil))
}


```

输出：
```
a name
b name
c name
a name
b name
c name
```

上面那个例子能看懂基本就掌握了inject的使用了。