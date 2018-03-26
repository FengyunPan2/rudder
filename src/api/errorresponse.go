package api

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
)

func handleInternalError(response *restful.Response, err error) {
	log.Printf("InternalError: %v",err)

	statusCode := http.StatusInternalServerError
	/*statusError, ok := err.(*errorsK8s.StatusError)
	if ok && statusError.Status().Code > 0 {
		statusCode = int(statusError.Status().Code)
	}*/

	response.AddHeader("Content-Type", "text/plain")
	response.WriteErrorString(statusCode, err.Error()+"\n")
}
