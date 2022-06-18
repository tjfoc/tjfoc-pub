package typeclass

type Read interface {
	Read(string) error
}
