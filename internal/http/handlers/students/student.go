package student

import (
	"encoding/json"
	"net/http"
	"pranayteja31/Urlshortener/internal/types"
)

//creation of handler in remote place

func New() http.HandlerFunc {
	return func (w http.ResponseWriter,r *http.Request)  {
		var student types.Student
		err := json.NewDecoder(r.Body).Decode(&student)
		if err != nil {
			
		}
	}
}
