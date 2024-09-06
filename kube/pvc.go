package kube

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PVCOptions struct {
	Name             string
	Namespace        string
	AccessMode       string `form:"accessmode" json:"accessmode"`
	StorageClassName string `form:"storageclassname" json:"storageclassname"`
	StorageSize      string `form:"storagesize" json:"storagesize"`
}

func CreateOrUpdatePVC(clientset *kubernetes.Clientset, ctx context.Context, opts PVCOptions, logHandler func(msg string)) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode(MustConvert(opts.AccessMode)),
			},
			StorageClassName: &opts.StorageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(opts.StorageSize),
				},
			},
		},
	}

	if _, err := clientset.CoreV1().PersistentVolumeClaims(opts.Namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create pvc resource: %v", err)
		}
		logHandler("pvc resource successfully updated")
	} else {
		logHandler("pvc resource successfully created")
	}

	return nil
}

func DeletePVC(clientset *kubernetes.Clientset, ctx context.Context, opts HPAOptions, logHandler func(msg string)) error {
	err := clientset.CoreV1().PersistentVolumeClaims(opts.Namespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete hpa resource: %v", err)
	}
	if apierrors.IsNotFound(err) {
		logHandler(fmt.Sprintf("pvc resource %s in namespace %s not found, no action taken\n", opts.Name, opts.Namespace))
	} else {
		logHandler(fmt.Sprintf("pvc resource %s in namespace %s successfully deleted\n", opts.Name, opts.Namespace))
	}
	return nil
}

func MustConvert(v string) corev1.PersistentVolumeAccessMode {
	v = strings.ToLower(v)
	switch v {
	case "readonlymany":
		return corev1.ReadOnlyMany
	case "readwriteonce":
		return corev1.ReadWriteOnce
	case "readwritemany":
		return corev1.ReadWriteMany
	default:
		panic(fmt.Errorf("unsupported access mode: %s", v))
	}
}
