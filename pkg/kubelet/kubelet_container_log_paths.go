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

package kubelet

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/kubelet/kuberuntime"
)

// GetContainerLogPaths returns a map of container names to their log directory
// paths for all containers (init, regular, and ephemeral) in the given pod.
func (kl *Kubelet) GetContainerLogPaths(_ context.Context, podNamespace, podName string) (map[string]string, error) {
	pod, ok := kl.GetPodByName(podNamespace, podName)
	if !ok {
		return nil, fmt.Errorf("pod %q in namespace %q not found", podName, podNamespace)
	}

	var podUID types.UID
	pod, mirrorPod, wasMirror := kl.podManager.GetPodAndMirrorPod(pod)
	if wasMirror {
		if pod == nil {
			return nil, fmt.Errorf("mirror pod %q does not have a corresponding pod", podName)
		}
		podUID = mirrorPod.UID
	} else {
		podUID = pod.UID
	}

	result := make(map[string]string)
	podLogsDir := kl.getPodLogsDir()

	for _, c := range pod.Spec.InitContainers {
		result[c.Name] = kuberuntime.BuildContainerLogsDirectory(podLogsDir, podNamespace, podName, podUID, c.Name)
	}
	for _, c := range pod.Spec.Containers {
		result[c.Name] = kuberuntime.BuildContainerLogsDirectory(podLogsDir, podNamespace, podName, podUID, c.Name)
	}
	for _, c := range pod.Spec.EphemeralContainers {
		result[c.Name] = kuberuntime.BuildContainerLogsDirectory(podLogsDir, podNamespace, podName, podUID, c.Name)
	}

	return result, nil
}
