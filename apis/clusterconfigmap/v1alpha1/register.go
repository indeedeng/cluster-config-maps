package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "indeed.com"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
	AddToScheme   = SchemeBuilder.AddToScheme
)

// cluster config map types metadata.
var (
	ClusterConfigMapKind             = reflect.TypeOf(ClusterConfigMap{}).Name()
	ClusterConfigMapGroupKind        = schema.GroupKind{Group: Group, Kind: ClusterConfigMapKind}.String()
	ClusterConfigMapKindAPIVersion   = ClusterConfigMapKind + "." + SchemeGroupVersion.String()
	ClusterConfigMapGroupVersionKind = SchemeGroupVersion.WithKind(ClusterConfigMapKind)
)

func init() {
	SchemeBuilder.Register(&ClusterConfigMap{}, &ClusterConfigMapList{})
}
