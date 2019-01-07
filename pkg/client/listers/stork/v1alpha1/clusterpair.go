/*
Copyright 2018 Openstorage.org

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterPairLister helps list ClusterPairs.
type ClusterPairLister interface {
	// List lists all ClusterPairs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterPair, err error)
	// ClusterPairs returns an object that can list and get ClusterPairs.
	ClusterPairs(namespace string) ClusterPairNamespaceLister
	ClusterPairListerExpansion
}

// clusterPairLister implements the ClusterPairLister interface.
type clusterPairLister struct {
	indexer cache.Indexer
}

// NewClusterPairLister returns a new ClusterPairLister.
func NewClusterPairLister(indexer cache.Indexer) ClusterPairLister {
	return &clusterPairLister{indexer: indexer}
}

// List lists all ClusterPairs in the indexer.
func (s *clusterPairLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterPair, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterPair))
	})
	return ret, err
}

// ClusterPairs returns an object that can list and get ClusterPairs.
func (s *clusterPairLister) ClusterPairs(namespace string) ClusterPairNamespaceLister {
	return clusterPairNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClusterPairNamespaceLister helps list and get ClusterPairs.
type ClusterPairNamespaceLister interface {
	// List lists all ClusterPairs in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterPair, err error)
	// Get retrieves the ClusterPair from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.ClusterPair, error)
	ClusterPairNamespaceListerExpansion
}

// clusterPairNamespaceLister implements the ClusterPairNamespaceLister
// interface.
type clusterPairNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClusterPairs in the indexer for a given namespace.
func (s clusterPairNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterPair, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterPair))
	})
	return ret, err
}

// Get retrieves the ClusterPair from the indexer for a given namespace and name.
func (s clusterPairNamespaceLister) Get(name string) (*v1alpha1.ClusterPair, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("clusterpair"), name)
	}
	return obj.(*v1alpha1.ClusterPair), nil
}