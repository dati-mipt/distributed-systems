package ftlregister

// Register определяет интерфейс
type Register interface {
	Write(value int64) bool
	Read() int64
}
