/*
Copyright 2015 The Kubernetes Authors.
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

package kubernetes

import (
	api "k8s.io/api/core/v1"
)

// IsPodReady returns true if a pod is ready; false otherwise.
// It returns an ErrPodCompleted error if the Pod is in api.PodFailed or api.PodSucceeded status phases.
func IsPodReady(pod *api.Pod) (bool, error) {
	if pod.Status.Phase == api.PodFailed || pod.Status.Phase == api.PodSucceeded {
		return false, ErrPodCompleted
	}
	return isPodReady(pod), nil
}

// IsPodSucceeded returns true if a pod is in the Succeeded state; false otherwise.
func IsPodSucceeded(pod *api.Pod) bool {
	return pod != nil && pod.Status.Phase == api.PodSucceeded
}

// IsPodFailed returns true if a pod is in the Failed state; false otherwise.
func IsPodFailed(pod *api.Pod) bool {
	return pod != nil && pod.Status.Phase == api.PodFailed
}

// isPodReady returns true if a pod is ready; false otherwise.
// Copied from: https://github.com/kubernetes/kubernetes/blob/master/pkg/api/pod/util.go#L237
func isPodReady(pod *api.Pod) bool {
	return isPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
// Copied from: https://github.com/kubernetes/kubernetes/blob/master/pkg/api/pod/util.go#L242
func isPodReadyConditionTrue(status api.PodStatus) bool {
	condition := getPodReadyCondition(status)
	return condition != nil && condition.Status == api.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
// Copied from: https://github.com/kubernetes/kubernetes/blob/master/pkg/api/pod/util.go#L249
func getPodReadyCondition(status api.PodStatus) *api.PodCondition {
	_, condition := getPodCondition(&status, api.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
// Copied from: https://github.com/kubernetes/kubernetes/blob/master/pkg/api/pod/util.go#L256
func getPodCondition(status *api.PodStatus, conditionType api.PodConditionType) (int, *api.PodCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}
