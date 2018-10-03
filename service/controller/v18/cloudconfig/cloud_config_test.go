package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_0_0"
	"github.com/giantswarm/legacycerts/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v18/encrypter"
)

func Test_Service_CloudConfig_NewMasterTemplate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		Certs        legacy.CompactTLSAssets
		ClusterKeys  randomkeys.Cluster
	}{
		{
			CustomObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Etcd: v1alpha1.ClusterEtcd{
							Port: 2379,
						},
					},
				},
			},
			Certs: legacy.CompactTLSAssets{
				CalicoClientCA:  "123456789-super-magic-calico-client-ca",
				CalicoClientCrt: "123456789-super-magic-calico-client-crt",
				CalicoClientKey: "123456789-super-magic-calico-client-key",
			},
			ClusterKeys: randomkeys.Cluster{
				APIServerEncryptionKey: randomkeys.RandomKey("fekhfiwoiqhoifhwqefoiqwefoikqhwef"),
			},
		},
	}

	for _, tc := range testCases {
		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		template, err := ccService.NewMasterTemplate(context.TODO(), tc.CustomObject, tc.Certs, tc.ClusterKeys)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decodedBytes := []byte(decoded)
		_, err = ignition.ConvertTemplatetoJSON(decodedBytes)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		calicoClientCA := []byte(tc.Certs.CalicoClientCA)
		t.Run("VerifyAPIServerCA", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientCA)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client CA", "none")
			}
		})

		calicoClientCrt := []byte(tc.Certs.CalicoClientCrt)
		t.Run("VerifyAPIServerCrt", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientCrt)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Crt", "none")
			}
		})

		calicoClientKey := []byte(tc.Certs.CalicoClientKey)
		t.Run("VerifyAPIServerKey", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientKey)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Key", "none")
			}
		})

		t.Run("VerifyTLSAssetsDecryptionUnit", func(t *testing.T) {
			if !strings.Contains(decoded, "decrypt-tls-assets.service") {
				t.Fatalf("expected %#v got %#v", "cloud config to contain unit decrypt-tls-assets.service", "none")
			}
		})

		serverEncryptionKey := []byte("H4sIAAAAAAAA/1SNMQ7CMAxF957CF+jQNSviCuwldYgVYTd2aBQh7o4CVREeLL33pf8T8eLgzF7bWkj4JBzoNswrXVCNhB1s06Bo8lCP5gaAEf6wC0OvWOxDq8pGC+oRzmj+6r/UL2GzH43A8x1dt9MhYW90EDDFQFUoR6EQa8YglGv/KceKYR+hBblQaQ6er3cAAAD//9QjGEbUAAAA")
		t.Run("VerifyAPIServerEncryptionKey", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(serverEncryptionKey)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain apiserver encryption config", "wrong config output")
			}
		})
	}
}

func Test_Service_CloudConfig_NewWorkerTemplate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		Certs        legacy.CompactTLSAssets
	}{
		{
			CustomObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "123456789-super-magic-aws-region",
					},
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			Certs: legacy.CompactTLSAssets{
				CalicoClientCA:  "123456789-super-magic-calico-client-ca",
				CalicoClientCrt: "123456789-super-magic-calico-client-crt",
				CalicoClientKey: "123456789-super-magic-calico-client-key",
			},
		},
	}

	for _, tc := range testCases {
		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		template, err := ccService.NewWorkerTemplate(context.TODO(), tc.CustomObject, tc.Certs)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		calicoClientCA := []byte(tc.Certs.CalicoClientCA)
		t.Run("VerifyAPIServerCA", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientCA)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client CA", "none")
			}
		})

		calicoClientCrt := []byte(tc.Certs.CalicoClientCrt)
		t.Run("VerifyAPIServerCrt", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientCrt)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Crt", "none")
			}
		})

		calicoClientKey := []byte(tc.Certs.CalicoClientKey)
		t.Run("VerifyAPIServerKey", func(t *testing.T) {
			if !strings.Contains(decoded, base64.StdEncoding.EncodeToString(calicoClientKey)) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Key", "none")
			}
		})

		t.Run("VerifyTLSAssetsDecryptionUnit", func(t *testing.T) {
			if !strings.Contains(decoded, "decrypt-tls-assets.service") {
				t.Fatalf("expected %#v got %#v", "cloud config to contain unit decrypt-tls-assets.service", "none")
			}
		})
	}
}

func testDecodeTemplate(template string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(template)
	if err != nil {
		return "", err
	}
	r, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, r)
	if err != nil {
		return "", err
	}
	r.Close()

	return b.String(), nil
}

func testNewCloudConfigService() (*CloudConfig, error) {
	var ccService *CloudConfig
	{
		packagePath, err := k8scloudconfig.GetPackagePath()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := Config{
			Encrypter:      &encrypter.EncrypterMock{},
			Logger:         microloggertest.New(),
			IgnitionPath:   packagePath,
			RegistryDomain: "quay.io",
		}

		ccService, err = New(c)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}
