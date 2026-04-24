package chi

import nethttp "net/http"

type Router interface {
	nethttp.Handler
	Get(pattern string, handlerFn nethttp.HandlerFunc)
}

type Mux struct {
	mux *nethttp.ServeMux
}

func NewRouter() *Mux {
	return &Mux{mux: nethttp.NewServeMux()}
}

func (m *Mux) Get(pattern string, handlerFn nethttp.HandlerFunc) {
	m.mux.HandleFunc(nethttp.MethodGet+" "+pattern, handlerFn)
}

func (m *Mux) ServeHTTP(w nethttp.ResponseWriter, r *nethttp.Request) {
	m.mux.ServeHTTP(w, r)
}
