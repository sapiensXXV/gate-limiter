package types

import "net/http"

type ProxyHandler interface {
	ToOrigin(w http.ResponseWriter, r *http.Request, origin string)
}
