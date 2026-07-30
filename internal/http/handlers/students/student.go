package student

import "net/http"

//creation of handler in remote place

func New() http.HandlerFunc {
	return func (w http.ResponseWriter,r *http.Request)  {
		w.Write([]byte("welcome to the api man"))
	}
}
