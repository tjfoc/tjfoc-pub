package functor

// MapFunc map a value to a new value
type MapFunc (func(interface{}) interface{})

// a function that analyze a recursive data structure and through use of a given combining operation
type FoldFunc (func(interface{}, interface{}) interface{})

// functional Functor
type Functor interface {
	Map(f MapFunc) Functor
	Foldl(f FoldFunc, a interface{}) interface{}
	Foldr(f FoldFunc, a interface{}) interface{}
}

/*
 * Name
 * 		Map
 * DESCRIPTION
 * 		functional Map
 * ARGUMENT
 * 		f - map a value to a new value
 * 		a - any value
 * RETURN VALUE
 * 		Map return a new value returned by MapFunc
 */
func Map(f MapFunc, a interface{}) interface{} {
	return f(a)
}

/*
 * Name
 * 		Foldl
 * DESCRIPTION
 * 		functional Foldl
 * AGRUMENt
 * 		f - a function that analyze a recursive data structure and through use of a given combining operation
 * 		a - any value
 * 		b - any value
 * RETURN VALUE
 * 		Foldl return a new value returned by Foldl
 */
func Foldl(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}

// the reverse order of foldl
func Foldr(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}
