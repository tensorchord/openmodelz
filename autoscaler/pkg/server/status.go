// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/autoscaler/pkg/version"
)

func getInfo(w http.ResponseWriter, r *http.Request) {
	scalerInfo := map[string]string{"version": version.GetEnvdVersion()}
	jsonOut, marshalErr := json.Marshal(scalerInfo)
	if marshalErr != nil {
		logrus.Infof("Error during unmarshal of autoscaler info request %s\n", marshalErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonOut)
}

func RunInfoServe() {
	tcpPort := 8080

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/system/info", getInfo)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        serverMux,
	}
	s.ListenAndServe()
}
