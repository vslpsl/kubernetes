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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestContainerLogPaths(t *testing.T) {
	fw := newServerTest()
	defer fw.testHTTPServer.Close()

	podNamespace := "other"
	podName := "foo"
	podUID := types.UID("test-uid-123")

	fw.fakeKubelet.podByNameFunc = func(namespace, name string) (*v1.Pod, bool) {
		if namespace == podNamespace && name == podName {
			return &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: podNamespace,
					Name:      podName,
					UID:       podUID,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{Name: "init-container"},
					},
					Containers: []v1.Container{
						{Name: "main-app"},
						{Name: "sidecar"},
					},
					EphemeralContainers: []v1.EphemeralContainer{
						{EphemeralContainerCommon: v1.EphemeralContainerCommon{Name: "debug"}},
					},
				},
			}, true
		}
		return nil, false
	}

	resp, err := http.Get(fw.testHTTPServer.URL + "/containerLogPaths/" + podNamespace + "/" + podName)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	expected := map[string]string{
		"init-container": fmt.Sprintf("/var/log/pods/%s_%s_%s/init-container", podNamespace, podName, podUID),
		"main-app":       fmt.Sprintf("/var/log/pods/%s_%s_%s/main-app", podNamespace, podName, podUID),
		"sidecar":        fmt.Sprintf("/var/log/pods/%s_%s_%s/sidecar", podNamespace, podName, podUID),
		"debug":          fmt.Sprintf("/var/log/pods/%s_%s_%s/debug", podNamespace, podName, podUID),
	}
	assert.Equal(t, expected, result)
}

func TestContainerLogPathsPodNotFound(t *testing.T) {
	fw := newServerTest()
	defer fw.testHTTPServer.Close()

	fw.fakeKubelet.podByNameFunc = func(namespace, name string) (*v1.Pod, bool) {
		return nil, false
	}

	resp, err := http.Get(fw.testHTTPServer.URL + "/containerLogPaths/default/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestContainerLogPathsNoContainers(t *testing.T) {
	fw := newServerTest()
	defer fw.testHTTPServer.Close()

	podNamespace := "default"
	podName := "empty-pod"
	podUID := types.UID("empty-uid")

	fw.fakeKubelet.podByNameFunc = func(namespace, name string) (*v1.Pod, bool) {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podNamespace,
				Name:      podName,
				UID:       podUID,
			},
			Spec: v1.PodSpec{},
		}, true
	}

	resp, err := http.Get(fw.testHTTPServer.URL + "/containerLogPaths/" + podNamespace + "/" + podName)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Empty(t, result)
}
