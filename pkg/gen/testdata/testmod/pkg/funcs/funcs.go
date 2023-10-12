package funcs

import (
	"io"

	corev1 "k8s.io/api/core/v1"
)

func foo(pod *corev1.Pod) string {
	return pod.Name
}

func Bar[T1 any, T2 int](t1 []T1, t2 T2) T2 {
	return t2
}

func Baz(pod *corev1.Pod, writer io.Writer, str string) error {
	return nil
}
