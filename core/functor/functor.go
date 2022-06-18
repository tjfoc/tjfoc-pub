package functor

type MapFunc (func(interface{}) interface{})
type FoldFunc (func(interface{}, interface{}) interface{})

type Functor interface {
	Map(f MapFunc) Functor
	Foldl(f FoldFunc, a interface{}) interface{}
	Foldr(f FoldFunc, a interface{}) interface{}
}

/*
 * MAP
 * 	功能:
 * 		映射
 * 	参数:
 * 		f为函数，用于将元素映射成一个新的元素
 * 		a
 * 	返回值:
 * 		返回一个新的元素
 */
func Map(f MapFunc, a interface{}) interface{} {
	return f(a)
}

/*
 * Foldl
 * 	功能:
 * 		左折，将一个元素映射成一个新的元素
 * 	参数:
 * 		f为函数，用于将元素映射成一个新的元素
 * 		a
 * 		b为新的容器
 * 返回值:
 *  	返回类型b类型的新元素
 */
func Foldl(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}

/*
 * Foldr
 * 	功能:
 * 		右折，将一个slice映射成一个新的元素
 * 	参数:
 * 		f为函数
 * 		a
 * 		b为新的容器
 * 返回值:
 *  	返回类型b类型的新元素
 */
func Foldr(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}
