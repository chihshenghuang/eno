//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Binding) DeepCopyInto(out *Binding) {
	*out = *in
	out.Resource = in.Resource
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Binding.
func (in *Binding) DeepCopy() *Binding {
	if in == nil {
		return nil
	}
	out := new(Binding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Composition) DeepCopyInto(out *Composition) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Composition.
func (in *Composition) DeepCopy() *Composition {
	if in == nil {
		return nil
	}
	out := new(Composition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Composition) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CompositionList) DeepCopyInto(out *CompositionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Composition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CompositionList.
func (in *CompositionList) DeepCopy() *CompositionList {
	if in == nil {
		return nil
	}
	out := new(CompositionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CompositionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CompositionSpec) DeepCopyInto(out *CompositionSpec) {
	*out = *in
	out.Synthesizer = in.Synthesizer
	if in.Bindings != nil {
		in, out := &in.Bindings, &out.Bindings
		*out = make([]Binding, len(*in))
		copy(*out, *in)
	}
	if in.SynthesisEnv != nil {
		in, out := &in.SynthesisEnv, &out.SynthesisEnv
		*out = make([]EnvVar, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CompositionSpec.
func (in *CompositionSpec) DeepCopy() *CompositionSpec {
	if in == nil {
		return nil
	}
	out := new(CompositionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CompositionStatus) DeepCopyInto(out *CompositionStatus) {
	*out = *in
	if in.Simplified != nil {
		in, out := &in.Simplified, &out.Simplified
		*out = new(SimplifiedStatus)
		**out = **in
	}
	if in.InFlightSynthesis != nil {
		in, out := &in.InFlightSynthesis, &out.InFlightSynthesis
		*out = new(Synthesis)
		(*in).DeepCopyInto(*out)
	}
	if in.CurrentSynthesis != nil {
		in, out := &in.CurrentSynthesis, &out.CurrentSynthesis
		*out = new(Synthesis)
		(*in).DeepCopyInto(*out)
	}
	if in.PreviousSynthesis != nil {
		in, out := &in.PreviousSynthesis, &out.PreviousSynthesis
		*out = new(Synthesis)
		(*in).DeepCopyInto(*out)
	}
	if in.InputRevisions != nil {
		in, out := &in.InputRevisions, &out.InputRevisions
		*out = make([]InputRevisions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PendingResynthesis != nil {
		in, out := &in.PendingResynthesis, &out.PendingResynthesis
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CompositionStatus.
func (in *CompositionStatus) DeepCopy() *CompositionStatus {
	if in == nil {
		return nil
	}
	out := new(CompositionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvVar) DeepCopyInto(out *EnvVar) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvVar.
func (in *EnvVar) DeepCopy() *EnvVar {
	if in == nil {
		return nil
	}
	out := new(EnvVar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Input) DeepCopyInto(out *Input) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.Resource = in.Resource
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Input.
func (in *Input) DeepCopy() *Input {
	if in == nil {
		return nil
	}
	out := new(Input)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputResource) DeepCopyInto(out *InputResource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputResource.
func (in *InputResource) DeepCopy() *InputResource {
	if in == nil {
		return nil
	}
	out := new(InputResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputRevisions) DeepCopyInto(out *InputRevisions) {
	*out = *in
	if in.Revision != nil {
		in, out := &in.Revision, &out.Revision
		*out = new(int)
		**out = **in
	}
	if in.SynthesizerGeneration != nil {
		in, out := &in.SynthesizerGeneration, &out.SynthesizerGeneration
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputRevisions.
func (in *InputRevisions) DeepCopy() *InputRevisions {
	if in == nil {
		return nil
	}
	out := new(InputRevisions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Manifest) DeepCopyInto(out *Manifest) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Manifest.
func (in *Manifest) DeepCopy() *Manifest {
	if in == nil {
		return nil
	}
	out := new(Manifest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodOverrides) DeepCopyInto(out *PodOverrides) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodOverrides.
func (in *PodOverrides) DeepCopy() *PodOverrides {
	if in == nil {
		return nil
	}
	out := new(PodOverrides)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ref) DeepCopyInto(out *Ref) {
	*out = *in
	out.Resource = in.Resource
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ref.
func (in *Ref) DeepCopy() *Ref {
	if in == nil {
		return nil
	}
	out := new(Ref)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceBinding) DeepCopyInto(out *ResourceBinding) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceBinding.
func (in *ResourceBinding) DeepCopy() *ResourceBinding {
	if in == nil {
		return nil
	}
	out := new(ResourceBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceRef) DeepCopyInto(out *ResourceRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceRef.
func (in *ResourceRef) DeepCopy() *ResourceRef {
	if in == nil {
		return nil
	}
	out := new(ResourceRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSlice) DeepCopyInto(out *ResourceSlice) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSlice.
func (in *ResourceSlice) DeepCopy() *ResourceSlice {
	if in == nil {
		return nil
	}
	out := new(ResourceSlice)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceSlice) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSliceList) DeepCopyInto(out *ResourceSliceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceSlice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSliceList.
func (in *ResourceSliceList) DeepCopy() *ResourceSliceList {
	if in == nil {
		return nil
	}
	out := new(ResourceSliceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceSliceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSliceRef) DeepCopyInto(out *ResourceSliceRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSliceRef.
func (in *ResourceSliceRef) DeepCopy() *ResourceSliceRef {
	if in == nil {
		return nil
	}
	out := new(ResourceSliceRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSliceSpec) DeepCopyInto(out *ResourceSliceSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]Manifest, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSliceSpec.
func (in *ResourceSliceSpec) DeepCopy() *ResourceSliceSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceSliceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSliceStatus) DeepCopyInto(out *ResourceSliceStatus) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceState, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSliceStatus.
func (in *ResourceSliceStatus) DeepCopy() *ResourceSliceStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceSliceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceState) DeepCopyInto(out *ResourceState) {
	*out = *in
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceState.
func (in *ResourceState) DeepCopy() *ResourceState {
	if in == nil {
		return nil
	}
	out := new(ResourceState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Result) DeepCopyInto(out *Result) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Result.
func (in *Result) DeepCopy() *Result {
	if in == nil {
		return nil
	}
	out := new(Result)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimplifiedStatus) DeepCopyInto(out *SimplifiedStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimplifiedStatus.
func (in *SimplifiedStatus) DeepCopy() *SimplifiedStatus {
	if in == nil {
		return nil
	}
	out := new(SimplifiedStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Symphony) DeepCopyInto(out *Symphony) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Symphony.
func (in *Symphony) DeepCopy() *Symphony {
	if in == nil {
		return nil
	}
	out := new(Symphony)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Symphony) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SymphonyList) DeepCopyInto(out *SymphonyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Symphony, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SymphonyList.
func (in *SymphonyList) DeepCopy() *SymphonyList {
	if in == nil {
		return nil
	}
	out := new(SymphonyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SymphonyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SymphonySpec) DeepCopyInto(out *SymphonySpec) {
	*out = *in
	if in.Variations != nil {
		in, out := &in.Variations, &out.Variations
		*out = make([]Variation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Bindings != nil {
		in, out := &in.Bindings, &out.Bindings
		*out = make([]Binding, len(*in))
		copy(*out, *in)
	}
	if in.SynthesisEnv != nil {
		in, out := &in.SynthesisEnv, &out.SynthesisEnv
		*out = make([]EnvVar, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SymphonySpec.
func (in *SymphonySpec) DeepCopy() *SymphonySpec {
	if in == nil {
		return nil
	}
	out := new(SymphonySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SymphonyStatus) DeepCopyInto(out *SymphonyStatus) {
	*out = *in
	if in.Synthesized != nil {
		in, out := &in.Synthesized, &out.Synthesized
		*out = (*in).DeepCopy()
	}
	if in.Reconciled != nil {
		in, out := &in.Reconciled, &out.Reconciled
		*out = (*in).DeepCopy()
	}
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = (*in).DeepCopy()
	}
	if in.Synthesizers != nil {
		in, out := &in.Synthesizers, &out.Synthesizers
		*out = make([]SynthesizerRef, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SymphonyStatus.
func (in *SymphonyStatus) DeepCopy() *SymphonyStatus {
	if in == nil {
		return nil
	}
	out := new(SymphonyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Synthesis) DeepCopyInto(out *Synthesis) {
	*out = *in
	if in.Initialized != nil {
		in, out := &in.Initialized, &out.Initialized
		*out = (*in).DeepCopy()
	}
	if in.PodCreation != nil {
		in, out := &in.PodCreation, &out.PodCreation
		*out = (*in).DeepCopy()
	}
	if in.Synthesized != nil {
		in, out := &in.Synthesized, &out.Synthesized
		*out = (*in).DeepCopy()
	}
	if in.Reconciled != nil {
		in, out := &in.Reconciled, &out.Reconciled
		*out = (*in).DeepCopy()
	}
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = (*in).DeepCopy()
	}
	if in.Canceled != nil {
		in, out := &in.Canceled, &out.Canceled
		*out = (*in).DeepCopy()
	}
	if in.ResourceSlices != nil {
		in, out := &in.ResourceSlices, &out.ResourceSlices
		*out = make([]*ResourceSliceRef, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ResourceSliceRef)
				**out = **in
			}
		}
	}
	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]Result, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InputRevisions != nil {
		in, out := &in.InputRevisions, &out.InputRevisions
		*out = make([]InputRevisions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Synthesis.
func (in *Synthesis) DeepCopy() *Synthesis {
	if in == nil {
		return nil
	}
	out := new(Synthesis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Synthesizer) DeepCopyInto(out *Synthesizer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Synthesizer.
func (in *Synthesizer) DeepCopy() *Synthesizer {
	if in == nil {
		return nil
	}
	out := new(Synthesizer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Synthesizer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SynthesizerList) DeepCopyInto(out *SynthesizerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Synthesizer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SynthesizerList.
func (in *SynthesizerList) DeepCopy() *SynthesizerList {
	if in == nil {
		return nil
	}
	out := new(SynthesizerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SynthesizerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SynthesizerRef) DeepCopyInto(out *SynthesizerRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SynthesizerRef.
func (in *SynthesizerRef) DeepCopy() *SynthesizerRef {
	if in == nil {
		return nil
	}
	out := new(SynthesizerRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SynthesizerSpec) DeepCopyInto(out *SynthesizerSpec) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExecTimeout != nil {
		in, out := &in.ExecTimeout, &out.ExecTimeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.PodTimeout != nil {
		in, out := &in.PodTimeout, &out.PodTimeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.Refs != nil {
		in, out := &in.Refs, &out.Refs
		*out = make([]Ref, len(*in))
		copy(*out, *in)
	}
	in.PodOverrides.DeepCopyInto(&out.PodOverrides)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SynthesizerSpec.
func (in *SynthesizerSpec) DeepCopy() *SynthesizerSpec {
	if in == nil {
		return nil
	}
	out := new(SynthesizerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SynthesizerStatus) DeepCopyInto(out *SynthesizerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SynthesizerStatus.
func (in *SynthesizerStatus) DeepCopy() *SynthesizerStatus {
	if in == nil {
		return nil
	}
	out := new(SynthesizerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Variation) DeepCopyInto(out *Variation) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Synthesizer = in.Synthesizer
	if in.Bindings != nil {
		in, out := &in.Bindings, &out.Bindings
		*out = make([]Binding, len(*in))
		copy(*out, *in)
	}
	if in.SynthesisEnv != nil {
		in, out := &in.SynthesisEnv, &out.SynthesisEnv
		*out = make([]EnvVar, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Variation.
func (in *Variation) DeepCopy() *Variation {
	if in == nil {
		return nil
	}
	out := new(Variation)
	in.DeepCopyInto(out)
	return out
}
