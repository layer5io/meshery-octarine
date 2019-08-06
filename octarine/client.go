// Copyright 2019 The Meshery Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package octarine

import (
	"time"

	"github.com/layer5io/meshery-octarine/meshes"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// OctarineClient represents an Octarine client in Meshery
type OctarineClient struct {
	config           *rest.Config
	k8sClientset     *kubernetes.Clientset
	k8sDynamicClient dynamic.Interface
	eventChan        chan *meshes.EventsResponse

	octarineAccount            string
	octarineControlPlane       string
	octarineAccMgrPword        string
	octarineDomain             string
	octarineCreatorPword       string
	octarineDeleterPword       string
	octarineReleaseVersion     string
	octarineReleaseDownloadURL string
	octarineReleaseUpdatedAt   time.Time
}

func configClient(kubeconfig []byte, contextName string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		ccfg, err := clientcmd.Load(kubeconfig)
		if err != nil {
			return nil, err
		}
		if contextName != "" {
			ccfg.CurrentContext = contextName
		}

		return clientcmd.NewDefaultClientConfig(*ccfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	}
	return rest.InClusterConfig()
}

func newClient(kubeconfig []byte, contextName string) (*OctarineClient, error) {
	client := OctarineClient{}
	config, err := configClient(kubeconfig, contextName)
	if err != nil {
		return nil, err
	}
	config.QPS = 100
	config.Burst = 200

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.k8sDynamicClient = dynamicClient

	k8sClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.k8sClientset = k8sClientset
	client.config = config

	return &client, nil
}
