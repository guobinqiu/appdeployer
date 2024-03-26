package resources

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressOptions struct {
	ApplicationName string
	Namespace       string
	Host            string
	TLS             bool
}

func CreateOrUpdateIngress(clientset *kubernetes.Clientset, ctx context.Context, opts IngressOptions) error {
	ingressClass := "nginx"
	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ApplicationName,
			Namespace: opts.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"nginx.ingress.kubernetes.io/ssl-passthrough":    "false",
				"nginx.ingress.kubernetes.io/backend-protocol":   "HTTP",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClass,
			Rules: []networkingv1.IngressRule{
				{
					Host: opts.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: opts.ApplicationName,
											Port: networkingv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if opts.TLS {
		if err := createOrUpdateTlsSecret(clientset, ctx, opts); err != nil {
			return err
		}
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: []string{
					opts.Host,
				},
				SecretName: "tls-" + opts.ApplicationName,
			},
		}
	}

	if _, err := clientset.NetworkingV1().Ingresses(opts.Namespace).Create(ctx, ingress, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	fmt.Println("kube ingress successfully done.")

	return nil
}

func createOrUpdateTlsSecret(clientset *kubernetes.Clientset, ctx context.Context, opts IngressOptions) error {
	// 创建CA证书和私钥
	caCert, caPrivateKey, err := createCACertificate()
	if err != nil {
		return err
	}

	// 创建服务器证书和私钥
	serverCertBytes, serverPrivateKey, err := createServerCertificate(caCert, caPrivateKey, opts.Host)
	if err != nil {
		return err
	}

	// 将证书和私钥保存到文件（PEM格式）
	tlsKeyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)})
	tlsCertBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})

	tlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tls-" + opts.ApplicationName,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.key": tlsKeyBytes,
			"tls.crt": tlsCertBytes,
		},
	}

	if _, err := clientset.CoreV1().Secrets(opts.Namespace).Create(ctx, tlsSecret, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			// return fmt.Errorf("failed to create TLS secret: %v", err)
			return err
		}
	}

	fmt.Println("kube tls secret successfully done.")

	return nil
}

// 创建一个CA
func createCACertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	// 生成私钥
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA private key: %v", err)
	}

	// 设置CA证书模板
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{""},
			OrganizationalUnit: []string{""},
			Country:            []string{""},
			Province:           []string{""},
			Locality:           []string{""},
			CommonName:         "",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 有效期10年
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// 根据模板创建自签名的CA证书
	caBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %v", err)
	}

	// 将CA证书解析为结构体
	caCert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse created CA certificate: %v", err)
	}

	return caCert, caPrivateKey, nil
}

// 使用CA签发服务器证书
func createServerCertificate(caCert *x509.Certificate, caPrivateKey *rsa.PrivateKey, serverName string) ([]byte, *rsa.PrivateKey, error) {
	// 生成服务器私钥
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate server private key: %v", err)
	}

	// 设置服务器证书模板
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{""},
			Country:      []string{""},
			Province:     []string{""},
			Locality:     []string{""},
			CommonName:   serverName,
		},
		DNSNames:    []string{serverName},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 有效期1年

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,

		Issuer: caCert.Subject,
	}

	// 根据CA签发服务器证书
	serverBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create server certificate: %v", err)
	}

	return serverBytes, serverPrivateKey, nil
}
