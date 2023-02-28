package common

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/giantswarm-e2e-tests/fixture"
)

// CommonTests adds common tests that we want to share between different providers.
func CommonTests(fixturePromise *fixture.Promise) {
	ctx := context.Background()

	It("nodes are all ready", func() {
		nodes := &corev1.NodeList{}
		err := fixturePromise.Cluster.GetWorkloadClusterKubeClient().List(ctx, nodes)
		Expect(err).ShouldNot(HaveOccurred())

		for _, node := range nodes.Items {
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" {
					Expect(condition.Status).Should(Equal(corev1.ConditionTrue))
				}
			}
		}
	})

	When("using PVCs", func() {
		var (
			err error
			pvc *corev1.PersistentVolumeClaim
			pod *corev1.Pod
		)

		BeforeEach(func() {
			pvc, err = createPVC(ctx, fixturePromise.Cluster.GetWorkloadClusterKubeClient(), "mypvc", "default")
			Expect(err).ShouldNot(HaveOccurred())

			pod, err = createPod(ctx, fixturePromise.Cluster.GetWorkloadClusterKubeClient(), "mypvc", "default")
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			err = fixturePromise.Cluster.GetWorkloadClusterKubeClient().Delete(ctx, pod)
			Expect(err).ShouldNot(HaveOccurred())

			err = fixturePromise.Cluster.GetWorkloadClusterKubeClient().Delete(ctx, pvc)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("supports using PVCs", func() {
			Eventually(func() error {
				current := &corev1.PersistentVolumeClaim{}
				err := fixturePromise.Cluster.GetWorkloadClusterKubeClient().Get(ctx, ctrl.ObjectKey{Name: pvc.Name, Namespace: pvc.Namespace}, current)
				Expect(err).ShouldNot(HaveOccurred())

				if current.Status.Phase != corev1.ClaimBound {
					return errors.New(fmt.Sprintf("PVC is not Bound, it's still in phase %s", current.Status.Phase))
				}

				return nil
			}, "5m", "5s").Should(Succeed())
		})
	})
}

func createPVC(ctx context.Context, ctrlClient ctrl.Client, pvcName, pvcNamespace string) (*corev1.PersistentVolumeClaim, error) {
	pvm := corev1.PersistentVolumeFilesystem

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: pvcNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			VolumeMode: &pvm,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse("8Gi"),
				},
			},
		},
	}
	err := ctrlClient.Create(ctx, pvc)
	if err != nil {
		return nil, err
	}

	return pvc, nil
}

func createPod(ctx context.Context, ctrlClient ctrl.Client, podName, podNamespace string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "mypod",
					Image: "quay.io/giantswarm/helloworld:latest",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "mypv",
							MountPath: "/mnt",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "mypv",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: podName,
						},
					},
				},
			},
		},
	}
	err := ctrlClient.Create(ctx, pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}
