/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

// getContainerLogPaths returns the log directory paths for all containers in the specified pod.
func (s *Server) getContainerLogPaths(request *restful.Request, response *restful.Response) {
	podNamespace := request.PathParameter("podNamespace")
	podID := request.PathParameter("podID")
	ctx := request.Request.Context()

	if len(podID) == 0 {
		response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "Missing podID."}`))
		return
	}
	if len(podNamespace) == 0 {
		response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "Missing podNamespace."}`))
		return
	}

	pod, ok := s.host.GetPodByName(podNamespace, podID)
	if !ok {
		response.WriteError(http.StatusNotFound, fmt.Errorf("pod %q does not exist", podID))
		return
	}

	logPaths, err := s.host.GetContainerLogPaths(ctx, pod.Namespace, pod.Name)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	data, err := json.Marshal(logPaths)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	writeJSONResponse(response, data)
}
