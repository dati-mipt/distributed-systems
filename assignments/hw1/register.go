package ftregister

type Register interface {
	Write(value int64) bool
	Read() int64
}
