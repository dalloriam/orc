package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"
	"path"

	"github.com/dalloriam/orc/orc"
	"github.com/spf13/viper"
)

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

func loadConfig() (orc.Config, error) {
	viper.SetConfigType("json")

	viper.AddConfigPath("./")

	homedir, err := getHomeDir()
	if err != nil {
		return orc.Config{}, err
	}

	viper.AddConfigPath(path.Join(homedir, ".config", "dalloriam"))

	viper.SetConfigName("orc")
	viper.ReadInConfig()

	cfg := orc.Config{}
	viper.Unmarshal(&cfg)

	return cfg, nil
}

func writeError(w io.Writer, errMsg string) {
	outBytes, _ := json.Marshal(map[string]string{"error": errMsg})
	w.Write(outBytes)
}

func registerAction(moduleName, actionName string, fn func(actionName string, data map[string]interface{}) ([]byte, error)) {
	http.HandleFunc(fmt.Sprintf("/%s/%s", moduleName, actionName),
		func(w http.ResponseWriter, r *http.Request) {
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

func main() {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	orcService, err := orc.New(cfg, registerAction)
	if err != nil {
		panic(err)
	}
	fmt.Println(orcService)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
