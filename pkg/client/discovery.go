package client

import (
	"net/url"

	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

// DiscoveryClient implements the functions that dicovery server-supported API groups,
// versions and resources.
type DiscoveryClient struct {
	*kclient.DiscoveryClient
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (d *DiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (resources *unversioned.APIResourceList, err error) {
	// we don't expose this version
	if groupVersion == "v1beta3" {
		return &unversioned.APIResourceList{}, nil
	}

	parentList, err := d.DiscoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return parentList, err
	}

	if groupVersion != "v1" {
		return parentList, nil
	}

	// we request v1, we must combine the parent list with the list from /oapi

	url := url.URL{}
	url.Path = "/oapi/" + groupVersion
	originResources := &unversioned.APIResourceList{}
	err = d.Get().AbsPath(url.String()).Do().Into(originResources)
	if err != nil {
		// ignore 403 or 404 error to be compatible with an v1.0 server.
		if groupVersion == "v1" && (errors.IsNotFound(err) || errors.IsForbidden(err)) {
			return parentList, nil
		} else {
			return nil, err
		}
	}

	parentList.APIResources = append(parentList.APIResources, originResources.APIResources...)
	return parentList, nil
}

// ServerResources returns the supported resources for all groups and versions.
func (d *DiscoveryClient) ServerResources() (map[string]*unversioned.APIResourceList, error) {
	apiGroups, err := d.ServerGroups()
	if err != nil {
		return nil, err
	}
	groupVersions := kclient.ExtractGroupVersions(apiGroups)
	result := map[string]*unversioned.APIResourceList{}
	for _, groupVersion := range groupVersions {
		resources, err := d.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}
		result[groupVersion] = resources
	}
	return result, nil
}

// New creates a new DiscoveryClient for the given RESTClient.
func NewDiscoveryClient(c *kclient.RESTClient) *DiscoveryClient {
	return &DiscoveryClient{kclient.NewDiscoveryClient(c)}
}
