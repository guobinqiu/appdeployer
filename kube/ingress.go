package kube

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
	"os"
	"path/filepath"
	"time"

	"github.com/guobinqiu/appdeployer/helpers"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressOptions struct {
	Name            string `form:"name" json:"name"`
	Namespace       string
	Host            string `form:"host" json:"host"`
	TLS             bool   `form:"tls" json:"tls"`
	SelfSigned      bool   `form:"selfsigned" json:"selfsigned"`
	SelfSignedYears int    `form:"selfsignedyears" json:"selfsignedyears"`
	CrtPath         string `form:"crtpath" json:"crtpath"`
	KeyPath         string `form:"keypath" json:"keypath"`
}

func CreateOrUpdateIngress(clientset *kubernetes.Clientset, ctx context.Context, opts IngressOptions, logHandler func(msg string)) error {
	ingressClass := "nginx"
	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
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
											Name: opts.Name,
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
		if err := CreateOrUpdateTlsSecret(clientset, ctx, opts); err != nil {
			return err
		}

		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: []string{
					opts.Host,
				},
				SecretName: "tls-" + opts.Name,
			},
		}
	}

	if _, err := clientset.NetworkingV1().Ingresses(opts.Namespace).Create(ctx, ingress, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ingress resource: %v", err)
		}
		logHandler("ingress resource successfully updated")
	} else {
		logHandler("ingress resource successfully created")
	}

	return nil
}

func CreateOrUpdateTlsSecret(clientset *kubernetes.Clientset, ctx context.Context, opts IngressOptions) error {
	var tlsKeyBytes, tlsCertBytes []byte

	if opts.SelfSigned {
		cm := &CertificateManager{}

		// 创建CA证书和私钥
		caCert, caPrivateKey, err := cm.CreateCACertificate(int(opts.SelfSignedYears))
		if err != nil {
			return fmt.Errorf("failed to create ca certificate: %v", err)
		}

		// 创建服务器证书和私钥
		serverCertBytes, serverPrivateKey, err := cm.CreateServerCertificate(caCert, caPrivateKey, opts.Host)
		if err != nil {
			return fmt.Errorf("failed to create server certificate: %v", err)
		}

		// 将证书和私钥保存到文件（PEM格式）
		tlsKeyBytes = cm.EncodePrivateKeyToPEM(serverPrivateKey)
		tlsCertBytes = cm.EncodeCertificateToPEM(serverCertBytes)
	} else {
		tlsCert, err := os.ReadFile(helpers.ExpandUser(filepath.Clean(opts.CrtPath)))
		if err != nil {
			return fmt.Errorf("failed to read certificate file: %v", err)
		}

		tlsKey, err := os.ReadFile(helpers.ExpandUser(filepath.Clean(opts.KeyPath)))
		if err != nil {
			return fmt.Errorf("failed to read key file: %v", err)
		}

		tlsKeyBytes = tlsCert
		tlsCertBytes = tlsKey
	}

	tlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tls-" + opts.Name,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSPrivateKeyKey: tlsKeyBytes,
			corev1.TLSCertKey:       tlsCertBytes,
		},
	}

	if _, err := clientset.CoreV1().Secrets(opts.Namespace).Create(ctx, tlsSecret, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create tls secret resource: %v", err)
		}
		fmt.Println("tls secret resource successfully updated")
	} else {
		fmt.Println("tls secret resource successfully created")
	}

	return nil
}

type CertificateManager struct{}

// 创建一个CA
func (cm *CertificateManager) CreateCACertificate(years int) (*x509.Certificate, *rsa.PrivateKey, error) {
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
		NotAfter:              time.Now().AddDate(years, 0, 0), // 有效期
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
func (cm *CertificateManager) CreateServerCertificate(caCert *x509.Certificate, caPrivateKey *rsa.PrivateKey, serverName string) ([]byte, *rsa.PrivateKey, error) {
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

// 将私钥转换为PEM格式
func (cm *CertificateManager) EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	return pem.EncodeToMemory(pemBlock)
}

// 将证书转换为PEM格式
func (cm *CertificateManager) EncodeCertificateToPEM(cert []byte) []byte {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}
	return pem.EncodeToMemory(pemBlock)
}
