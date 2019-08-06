// Copyright 2019 Layer5.io
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
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
	"github.com/layer5io/meshery-octarine/meshes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (oClient *OctarineClient) CreateMeshInstance(_ context.Context, k8sReq *meshes.CreateMeshInstanceRequest) (*meshes.CreateMeshInstanceResponse, error) {
	var k8sConfig []byte
	contextName := ""
	if k8sReq != nil {
		k8sConfig = k8sReq.K8SConfig
		contextName = k8sReq.ContextName
	}
	// logrus.Debugf("received k8sConfig: %s", k8sConfig)
	logrus.Debugf("received contextName: %s", contextName)

	oc, err := newClient(k8sConfig, contextName)
	if err != nil {
		err = errors.Wrapf(err, "unable to create a new Octarine client")
		logrus.Error(err)
		return nil, err
	}
	oClient.k8sClientset = oc.k8sClientset
	oClient.k8sDynamicClient = oc.k8sDynamicClient
	oClient.eventChan = make(chan *meshes.EventsResponse, 100)
	oClient.config = oc.config
	return &meshes.CreateMeshInstanceResponse{}, nil
}

func (oClient *OctarineClient) createResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	_, err := oClient.k8sDynamicClient.Resource(res).Namespace(data.GetNamespace()).Create(data, metav1.CreateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "unable to create the requested resource, attempting operation without namespace")
		logrus.Warn(err)
		_, err = oClient.k8sDynamicClient.Resource(res).Create(data, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "unable to create the requested resource, attempting to update")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Created Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

func (oClient *OctarineClient) deleteResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	if oClient.k8sDynamicClient == nil {
		return errors.New("mesh client has not been created")
	}

	if res.Resource == "namespaces" && data.GetName() == "default" { // skipping deletion of default namespace
		return nil
	}

	// in the case with deployments, have to scale it down to 0 first and then delete. . . or else RS and pods will be left behind
	if res.Resource == "deployments" {
		data1, err := oClient.getResource(ctx, res, data)
		if err != nil {
			return err
		}
		depl := data1.UnstructuredContent()
		spec1 := depl["spec"].(map[string]interface{})
		spec1["replicas"] = 0
		data1.SetUnstructuredContent(depl)
		if err = oClient.updateResource(ctx, res, data1); err != nil {
			return err
		}
	}
	policy := metav1.DeletePropagationBackground
	err := oClient.k8sDynamicClient.Resource(res).Namespace(data.GetNamespace()).Delete(data.GetName(),
	    &metav1.DeleteOptions{PropagationPolicy: &policy})
	if err != nil {
		err = errors.Wrapf(err, "unable to delete the requested resource, attempting operation without namespace")
		logrus.Warn(err)

		err := oClient.k8sDynamicClient.Resource(res).Delete(data.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			err = errors.Wrapf(err, "unable to delete the requested resource")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Deleted Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

func (oClient *OctarineClient) getResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	data1, err := oClient.k8sDynamicClient.Resource(res).Namespace(data.GetNamespace()).Get(data.GetName(), metav1.GetOptions{})
	if err != nil {
		err = errors.Wrap(err, "unable to retrieve the resource with a matching name, attempting operation without namespace")
		logrus.Warn(err)

		data1, err = oClient.k8sDynamicClient.Resource(res).Get(data.GetName(), metav1.GetOptions{})
		if err != nil {
			err = errors.Wrap(err, "unable to retrieve the resource with a matching name, while attempting to apply the config")
			logrus.Error(err)
			return nil, err
		}
	}
	logrus.Infof("Retrieved Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return data1, nil
}

func (oClient *OctarineClient) updateResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	if _, err := oClient.k8sDynamicClient.Resource(res).Namespace(data.GetNamespace()).Update(data, metav1.UpdateOptions{}); err != nil {
		err = errors.Wrap(err, "unable to update resource with the given name, attempting operation without namespace")
		logrus.Warn(err)

		if _, err = oClient.k8sDynamicClient.Resource(res).Update(data, metav1.UpdateOptions{}); err != nil {
			err = errors.Wrap(err, "unable to update resource with the given name, while attempting to apply the config")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Updated Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

// MeshName just returns the name of the mesh the client is representing
func (oClient *OctarineClient) MeshName(context.Context, *meshes.MeshNameRequest) (*meshes.MeshNameResponse, error) {
	return &meshes.MeshNameResponse{Name: "Octarine"}, nil
}

func (oClient *OctarineClient) applyManifestPayload(ctx context.Context, namespace string, newBytes []byte, delete bool) error {
	if oClient.k8sDynamicClient == nil {
		return errors.New("mesh client has not been created")
	}
	// logrus.Debugf("received yaml bytes: %s", newBytes)
	jsonBytes, err := yaml.YAMLToJSON(newBytes)
	if err != nil {
		err = errors.Wrapf(err, "unable to convert yaml to json")
		logrus.Error(err)
		return err
	}
	// logrus.Debugf("created json: %s, length: %d", jsonBytes, len(jsonBytes))
	if len(jsonBytes) > 5 { // attempting to skip 'null' json
		data := &unstructured.Unstructured{}
		err = data.UnmarshalJSON(jsonBytes)
		if err != nil {
			err = errors.Wrapf(err, "unable to unmarshal json created from yaml")
			logrus.Error(err)
			return err
		}
		if data.IsList() {
			err = data.EachListItem(func(r runtime.Object) error {
				dataL, _ := r.(*unstructured.Unstructured)
				return oClient.executeManifest(ctx, dataL, namespace, delete)
			})
			return err
		}
		return oClient.executeManifest(ctx, data, namespace, delete)
	}
	return nil
}

func (oClient *OctarineClient) executeManifest(ctx context.Context, data *unstructured.Unstructured, namespace string, delete bool) error {
	// logrus.Debug("========================================================")
	// logrus.Debugf("Received data: %+#v", data)
	if namespace != "" {
		data.SetNamespace(namespace)
	}
	groupVersion := strings.Split(data.GetAPIVersion(), "/")
	logrus.Debugf("groupVersion: %v", groupVersion)
	var group, version string
	if len(groupVersion) == 2 {
		group = groupVersion[0]
		version = groupVersion[1]
	} else if len(groupVersion) == 1 {
		version = groupVersion[0]
	}

	kind := strings.ToLower(data.GetKind())
	switch kind {
	case "logentry":
		kind = "logentries"
	case "kubernetes":
		kind = "kuberneteses"
	default:
		kind += "s"
	}

	res := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}
	logrus.Debugf("Computed Resource: %+#v", res)

	if delete {
		return oClient.deleteResource(ctx, res, data)
	}

	if err := oClient.createResource(ctx, res, data); err != nil {
		data1, err := oClient.getResource(ctx, res, data)
		if err != nil {
			return err
		}
		if err = oClient.updateResource(ctx, res, data1); err != nil {
			return err
		}
	}
	return nil
}

func (oClient *OctarineClient) labelNamespaceForAutoInjection(ctx context.Context, namespace string) error {
	ns := &unstructured.Unstructured{}
	res := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
	ns.SetName(namespace)
	ns, err := oClient.getResource(ctx, res, ns)
	if err != nil {
		return err
	}
	ns.SetLabels(map[string]string{
		"octarine-injection": "enabled",
	})
	err = oClient.updateResource(ctx, res, ns)
	if err != nil {
		return err
	}
	secret := &unstructured.Unstructured{}
	res = schema.GroupVersionResource{
		Version:  "v1",
		Resource: "secrets",
	}
	secret.SetName("docker-registry-secret")
	secret.SetNamespace(oClient.octarineDataplaneNs)
	secret, err = oClient.getResource(ctx, res, secret)
	if err != nil {
		return err
	}
	secret.SetNamespace(namespace)
	secret.SetResourceVersion("")
	err = oClient.createResource(ctx, res, secret)
	if err != nil {
		return err
	}
	return nil
}

func (oClient *OctarineClient) executeInstall(ctx context.Context, arReq *meshes.ApplyRuleRequest) error {
	if arReq.Namespace == "" {
		arReq.Namespace = "octarine-dataplane"
	}
	oClient.octarineDataplaneNs = arReq.Namespace
	if arReq.DeleteOp {
		defer oClient.deleteCpObjects()
	} else {
		if err := oClient.createCpObjects(); err != nil {
			return err
		}
	}
	dataplaneYaml, err := oClient.getOctarineYAMLs(arReq.Namespace)
	if err != nil {
		return err
	}
	if err := oClient.applyConfigChange(ctx, dataplaneYaml, arReq.Namespace, arReq.DeleteOp); err != nil {
		return err
	}
	return nil
}

func (oClient *OctarineClient) executeBookInfoInstall(ctx context.Context, arReq *meshes.ApplyRuleRequest) error {
	if !arReq.DeleteOp {
		if err := oClient.labelNamespaceForAutoInjection(ctx, arReq.Namespace); err != nil {
			return err
		}
	}
	yamlFileContents, err := oClient.getBookInfoAppYAML()
	if err != nil {
		return err
	}
	if err := oClient.applyConfigChange(ctx, yamlFileContents, arReq.Namespace, arReq.DeleteOp); err != nil {
		return err
	}
	return nil
}

// ApplyOperation is a method invoked to apply a particular operation on the mesh in a namespace
func (oClient *OctarineClient) ApplyOperation(ctx context.Context, arReq *meshes.ApplyRuleRequest) (*meshes.ApplyRuleResponse, error) {
	if arReq == nil {
		return nil, errors.New("mesh client has not been created")
	}

	op, ok := supportedOps[arReq.OpName]
	if !ok {
		return nil, fmt.Errorf("error: %s is not a valid operation name", arReq.OpName)
	}

	if arReq.OpName == customOpCommand && arReq.CustomBody == "" {
		return nil, fmt.Errorf("error: yaml body is empty for %s operation", arReq.OpName)
	}

	var yamlFileContents string
	// var err error

	switch arReq.OpName {
	case customOpCommand:
		yamlFileContents = arReq.CustomBody
	case installOctarineCommand:
		go func() {
			opName1 := "deploying"
			if arReq.DeleteOp {
				opName1 = "removing"
			}
			if err := oClient.executeInstall(ctx, arReq); err != nil {
				oClient.eventChan <- &meshes.EventsResponse{
					EventType: meshes.EventType_ERROR,
					Summary:   fmt.Sprintf("Error while %s Octarine", opName1),
					Details:   err.Error(),
				}
				return
			}
			opName := "deployed"
			if arReq.DeleteOp {
				opName = "removed"
			}
			oClient.eventChan <- &meshes.EventsResponse{
				EventType: meshes.EventType_INFO,
				Summary:   fmt.Sprintf("Octarine %s successfully", opName),
				Details:   fmt.Sprintf("The latest version of Octarine is now %s.", opName),
			}
			return
		}()
		return &meshes.ApplyRuleResponse{}, nil
	case installBookInfoCommand:
		go func() {
			opName1 := "deploying"
			if arReq.DeleteOp {
				opName1 = "removing"
			}
			if err := oClient.executeBookInfoInstall(ctx, arReq); err != nil {
				oClient.eventChan <- &meshes.EventsResponse{
					EventType: meshes.EventType_ERROR,
					Summary:   fmt.Sprintf("Error while %s the canonical Book Info App", opName1),
					Details:   err.Error(),
				}
				return
			}
			opName := "deployed"
			if arReq.DeleteOp {
				opName = "removed"
			}
			oClient.eventChan <- &meshes.EventsResponse{
				EventType: meshes.EventType_INFO,
				Summary:   fmt.Sprintf("Book Info app %s successfully", opName),
				Details:   fmt.Sprintf("The canonical Book Info app is now %s.", opName),
			}
			return
		}()
		return &meshes.ApplyRuleResponse{}, nil
	case runVet:
		go oClient.runVet()
		return &meshes.ApplyRuleResponse{}, nil
	default:
		tmpl, err := template.ParseFiles(path.Join("octarine", "config_templates", op.templateName))
		if err != nil {
			err = errors.Wrapf(err, "unable to parse template")
			logrus.Error(err)
			return nil, err
		}
		buf := bytes.NewBufferString("")
		err = tmpl.Execute(buf, map[string]string{
			"user_name": arReq.Username,
			"namespace": arReq.Namespace,
		})
		if err != nil {
			err = errors.Wrapf(err, "unable to execute template")
			logrus.Error(err)
			return nil, err
		}
		yamlFileContents = buf.String()
	}

	if err := oClient.applyConfigChange(ctx, yamlFileContents, arReq.Namespace, arReq.DeleteOp); err != nil {
		return nil, err
	}

	return &meshes.ApplyRuleResponse{}, nil
}

func (oClient *OctarineClient) applyConfigChange(ctx context.Context, yamlFileContents, namespace string, delete bool) error {
	yamls := strings.Split(yamlFileContents, "---")

	for _, yml := range yamls {
		if strings.TrimSpace(yml) != "" {
			if err := oClient.applyManifestPayload(ctx, namespace, []byte(yml), delete); err != nil {
				errStr := strings.TrimSpace(err.Error())
				if delete && (strings.HasSuffix(errStr, "not found") ||
					strings.HasSuffix(errStr, "the server could not find the requested resource")) {
					// logrus.Debugf("skipping error. . .")
					continue
				}
				// logrus.Debugf("returning error: %v", err)
				return err
			}
		}
	}
	return nil
}

// SupportedOperations - returns a list of supported operations on the mesh
func (oClient *OctarineClient) SupportedOperations(context.Context, *meshes.SupportedOperationsRequest) (*meshes.SupportedOperationsResponse, error) {
	result := map[string]string{}
	for key, op := range supportedOps {
		result[key] = op.name
	}
	return &meshes.SupportedOperationsResponse{
		Ops: result,
	}, nil
}

// StreamEvents - streams generated/collected events to the client
func (oClient *OctarineClient) StreamEvents(in *meshes.EventsRequest, stream meshes.MeshService_StreamEventsServer) error {
	logrus.Debugf("waiting on event stream. . .")
	for {
		select {
		case event := <-oClient.eventChan:
			logrus.Debugf("sending event: %+#v", event)
			if err := stream.Send(event); err != nil {
				err = errors.Wrapf(err, "unable to send event")

				// to prevent loosing the event, will re-add to the channel
				go func() {
					oClient.eventChan <- event
				}()
				logrus.Error(err)
				return err
			}
		default:
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
