package shops

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "shop.devops.io", Version: "v1alpha1"}
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &Shop{}, &ShopList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

type Shop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ShopSpec   `json:"spec,omitempty"`
	Status            ShopStatus `json:"status,omitempty"`
}

type ShopSpec struct {
	Availability  string `json:"availability,omitempty"`
	WalletAddress string `json:"walletAddress"`
	Database      string `json:"database,omitempty"`
	Image         string `json:"image,omitempty"`
}

type ShopStatus struct {
	Phase         string `json:"phase,omitempty"`
	Replicas      int32  `json:"replicas,omitempty"`
	ReadyReplicas int32  `json:"readyReplicas,omitempty"`
	DatabaseReady bool   `json:"databaseReady,omitempty"`
	ServiceURL    string `json:"serviceUrl,omitempty"`
}

func (s *Shop) DeepCopyObject() runtime.Object { c := s.DeepCopy(); return c }
func (s *Shop) DeepCopy() *Shop {
	if s == nil {
		return nil
	}
	out := new(Shop)
	*out = *s
	s.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = s.Spec
	out.Status = s.Status
	return out
}

type ShopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Shop `json:"items"`
}

func (sl *ShopList) DeepCopyObject() runtime.Object { return sl.DeepCopy() }
func (sl *ShopList) DeepCopy() *ShopList {
	if sl == nil {
		return nil
	}
	out := new(ShopList)
	*out = *sl
	sl.ListMeta.DeepCopyInto(&out.ListMeta)
	if sl.Items != nil {
		out.Items = make([]Shop, len(sl.Items))
		for i := range sl.Items {
			sl.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
	return out
}

func (s *Shop) DeepCopyInto(out *Shop) {
	*out = *s
	s.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = s.Spec
	out.Status = s.Status
}
