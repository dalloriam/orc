package interfaces

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

func writeError(w http.ResponseWriter, errMsg string) {
	outBytes, _ := json.Marshal(map[string]string{"error": errMsg})
	w.Write(outBytes)
	w.WriteHeader(http.StatusInternalServerError)
}

// HandleWithHTTP creates a HTTP handler for the action.
func HandleWithHTTP(moduleName, actionName string, fn func(actionName string, data map[string]interface{}) ([]byte, error)) {
	pattern := fmt.Sprintf("/%s/%s", moduleName, actionName)
	ctxLogger := logrus.WithFields(logrus.Fields{
		"module": moduleName,
		"action": actionName,
	})
	http.HandleFunc(pattern,
		func(w http.ResponseWriter, r *http.Request) {
            ctxLogger.Infof("received http request: %s", pattern)
			w.Header().Add("Content-Type", "application/json")
			// Read the data from the request
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, err.Error())
				return
			}

			// Decode the request into some dict.
			var parsed map[string]interface{}
			if len(body) > 0 {
				if err := json.Unmarshal(body, &parsed); err != nil {
					writeError(w, err.Error())
					return
				}
				outBytes, _ := json.Marshal(parsed)
				ctxLogger.Debugf("action payload: %s", string(outBytes))
			}

			// Fetch the response from the module & return the output.
			outBytes, err := fn(actionName, parsed)
			if err != nil {
				if len(outBytes) > 0 {
					writeError(w, string(outBytes))
				} else {
					writeError(w, err.Error())
				}
				return
			}

			w.Write(outBytes)
		})
}
