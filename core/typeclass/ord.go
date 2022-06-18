package typeclass

type Ord interface {
	LessThan(Ord) bool
	MoreThan(Ord) bool
}
