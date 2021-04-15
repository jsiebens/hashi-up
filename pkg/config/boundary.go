package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type BoundaryConfig struct {
	WorkerName           string
	ControllerName       string
	DatabaseURL          string
	RootKey              string
	WorkerAuthKey        string
	RecoveryKey          string
	ApiAddress           string
	ApiKeyFile           string
	ApiCertFile          string
	ClusterAddress       string
	ClusterKeyFile       string
	ClusterCertFile      string
	ProxyAddress         string
	ProxyKeyFile         string
	ProxyCertFile        string
	PublicAddress        string
	PublicClusterAddress string
	Controllers          []string
}

func (c *BoundaryConfig) HasDatabaseURL() bool {
	return len(c.DatabaseURL) != 0
}

func (c *BoundaryConfig) IsWorkerEnabled() bool {
	return len(c.WorkerName) != 0
}

func (c *BoundaryConfig) IsControllerEnabled() bool {
	return len(c.ControllerName) != 0
}

func (c *BoundaryConfig) HasAllRequiredControllerKeys() bool {
	return len(c.RootKey) != 0 && len(c.WorkerAuthKey) != 0 && len(c.RecoveryKey) != 0
}

func (c *BoundaryConfig) HasAllRequiredWorkerKeys() bool {
	return len(c.WorkerAuthKey) != 0
}

func (c *BoundaryConfig) HasValidApiTLSSettings() bool {
	return allOrNothing(c.ApiKeyFile, c.ApiCertFile)
}

func (c *BoundaryConfig) HasValidProxyTLSSettings() bool {
	return allOrNothing(c.ProxyKeyFile, c.ProxyCertFile)
}

func (c *BoundaryConfig) HasValidClusterTLSSettings() bool {
	return allOrNothing(c.ClusterKeyFile, c.ClusterCertFile)
}

func (c *BoundaryConfig) ApiTLSEnabled() bool {
	return len(c.ApiCertFile) != 0 && len(c.ApiKeyFile) != 0
}

func (c *BoundaryConfig) ProxyTLSEnabled() bool {
	return len(c.ProxyCertFile) != 0 && len(c.ProxyKeyFile) != 0
}

func (c *BoundaryConfig) ClusterTLSEnabled() bool {
	return len(c.ClusterCertFile) != 0 && len(c.ClusterKeyFile) != 0
}

func (c *BoundaryConfig) GenerateConfigFile() string {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	if len(c.ControllerName) != 0 {
		controllerBlock := rootBody.AppendNewBlock("controller", []string{})
		controllerBlock.Body().SetAttributeValue("name", cty.StringVal(c.ControllerName))

		databaseBlock := controllerBlock.Body().AppendNewBlock("database", []string{})
		databaseBlock.Body().SetAttributeValue("url", cty.StringVal(c.DatabaseURL))

		if len(c.PublicClusterAddress) != 0 {
			controllerBlock.Body().SetAttributeValue("public_cluster_addr", cty.StringVal(c.PublicClusterAddress))
		}
	}

	if len(c.WorkerName) != 0 {
		workerBlock := rootBody.AppendNewBlock("worker", []string{})
		workerBlock.Body().SetAttributeValue("name", cty.StringVal(c.WorkerName))
		workerBlock.Body().SetAttributeValue("controllers", cty.ListVal(transform(c.Controllers)))

		if len(c.PublicAddress) != 0 {
			workerBlock.Body().SetAttributeValue("public_addr", cty.StringVal(c.PublicAddress))
		}
	}

	if len(c.ControllerName) != 0 && len(c.ApiAddress) != 0 {
		apiAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		apiAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("api"))
		apiAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ApiAddress))

		if c.ApiTLSEnabled() {
			apiAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(false))
			apiAddressBlock.Body().SetAttributeValue("tls_cert_file", cty.StringVal(makeAbsolute(c.ApiCertFile, "/etc/boundary.d")))
			apiAddressBlock.Body().SetAttributeValue("tls_key_file", cty.StringVal(makeAbsolute(c.ApiKeyFile, "/etc/boundary.d")))
		} else {
			apiAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
		}
	}

	if len(c.ControllerName) != 0 && len(c.ClusterAddress) != 0 {
		clusterAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		clusterAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("cluster"))
		clusterAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ClusterAddress))
		clusterAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))

		if c.ClusterTLSEnabled() {
			clusterAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(false))
			clusterAddressBlock.Body().SetAttributeValue("tls_cert_file", cty.StringVal(makeAbsolute(c.ClusterCertFile, "/etc/boundary.d")))
			clusterAddressBlock.Body().SetAttributeValue("tls_key_file", cty.StringVal(makeAbsolute(c.ClusterKeyFile, "/etc/boundary.d")))
		} else {
			clusterAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
		}
	}

	if len(c.WorkerName) != 0 && len(c.ProxyAddress) != 0 {
		proxyAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		proxyAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("proxy"))
		proxyAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ProxyAddress))

		if c.ProxyTLSEnabled() {
			proxyAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(false))
			proxyAddressBlock.Body().SetAttributeValue("tls_cert_file", cty.StringVal(makeAbsolute(c.ProxyCertFile, "/etc/boundary.d")))
			proxyAddressBlock.Body().SetAttributeValue("tls_key_file", cty.StringVal(makeAbsolute(c.ProxyKeyFile, "/etc/boundary.d")))
		} else {
			proxyAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
		}
	}

	if len(c.RootKey) != 0 {
		rootKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		rootKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("root"))
		rootKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		rootKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.RootKey))
		rootKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_root"))
	}

	if len(c.WorkerAuthKey) != 0 {
		workerAuthKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		workerAuthKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("worker-auth"))
		workerAuthKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		workerAuthKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.WorkerAuthKey))
		workerAuthKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_worker-auth"))
	}

	if len(c.RecoveryKey) != 0 {
		recoveryKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		recoveryKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("recovery"))
		recoveryKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		recoveryKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.RecoveryKey))
		recoveryKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_recovery"))
	}

	return string(f.Bytes())
}

func allOrNothing(a, b string) bool {
	return (len(a) != 0 && len(b) != 0) || (len(a) == 0 && len(b) == 0)
}
