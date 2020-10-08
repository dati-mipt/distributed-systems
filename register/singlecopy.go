package register

type Server interface {
	WriteReq(value int64) bool
	ReadReq() int64
}

type SingleCopyRegisterClient struct {
	s Server
}

func (c SingleCopyRegisterClient) Write(value int64) bool {
	return c.s.WriteReq(value)
}

func (c SingleCopyRegisterClient) Read() int64 {
	return c.s.ReadReq()
}

type SingleCopyRegisterServer struct {
	current int64
}

func (s SingleCopyRegisterServer) WriteReq(value int64) bool {
	s.current = value
	return true
}

func (s SingleCopyRegisterServer) ReadReq() int64 {
	return s.current
}
