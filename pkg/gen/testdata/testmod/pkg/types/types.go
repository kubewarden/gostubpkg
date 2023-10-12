package types

import (
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
)

type MyStruct struct {
	MyEmbeddedStruct
	corev1.PodSpec
	Name     string
	Num      int
	Pointer  *os.File
	IOReader io.Reader
	Pod      corev1.Pod
}

type MyEmbeddedStruct struct{}

//go:generate mockery --name MyInterface
type MyInterface interface {
	GetPodName(pod *corev1.Pod) string
	getPodNamePrivate(pod *corev1.Pod) string
}

type (
	MyTypeAlias  string
	MyTypeAlias2 MyStruct
	MyTypeAlias3 corev1.Pod
)

func (s *MyStruct) GetPodName(pod *corev1.Pod) string {
	return pod.Name
}

func (s *MyStruct) getPodNamePrivate(pod *corev1.Pod) string {
	return pod.Name
}
