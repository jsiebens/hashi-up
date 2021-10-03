package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type VaultConfig struct {
	ApiAddr        string
	ClusterAddr    string
	Address        []string
	CertFile       string
	KeyFile        string
	Storage        string
	ConsulAddr     string
	ConsulPath     string
	ConsulToken    string
	ConsulCaFile   string
	ConsulCertFile string
	ConsulKeyFile  string
}

func (c VaultConfig) EnableTLS() bool {
	return len(c.CertFile) != 0 && len(c.KeyFile) != 0
}

func (c VaultConfig) EnableConsulTLS() bool {
	return len(c.ConsulCaFile) != 0 && len(c.ConsulCertFile) != 0 && len(c.ConsulKeyFile) != 0
}

func (c VaultConfig) GenerateConfigFile() string {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.SetAttributeValue("ui", cty.BoolVal(true))

	storageBlock := rootBody.AppendNewBlock("storage", []string{c.Storage})

	if c.Storage == "file" {
		storageBlock.Body().SetAttributeValue("path", cty.StringVal("/opt/vault"))
	}

	if c.Storage == "consul" {
		storageBlock.Body().SetAttributeValue("address", cty.StringVal(c.ConsulAddr))
		storageBlock.Body().SetAttributeValue("path", cty.StringVal(c.ConsulPath))

		if len(c.ConsulToken) != 0 {
			storageBlock.Body().SetAttributeValue("token", cty.StringVal(c.ConsulToken))
		}

		if c.EnableConsulTLS() {
			storageBlock.Body().SetAttributeValue("scheme", cty.StringVal("https"))
			storageBlock.Body().SetAttributeValue("tls_ca_file", cty.StringVal(makeAbsolute(c.ConsulCaFile, "/etc/vault.d")))
			storageBlock.Body().SetAttributeValue("tls_cert_file", cty.StringVal(makeAbsolute(c.ConsulCertFile, "/etc/vault.d")))
			storageBlock.Body().SetAttributeValue("tls_key_file", cty.StringVal(makeAbsolute(c.ConsulKeyFile, "/etc/vault.d")))
		}
	}

	if len(c.ApiAddr) != 0 {
		rootBody.SetAttributeValue("api_addr", cty.StringVal(c.ApiAddr))
	}

	if len(c.ClusterAddr) != 0 {
		rootBody.SetAttributeValue("cluster_addr", cty.StringVal(c.ClusterAddr))
	}

	for _, a := range c.Address {
		listenerBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		listenerBlock.Body().SetAttributeValue("address", cty.StringVal(a))

		if c.EnableTLS() {
			listenerBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(false))
			listenerBlock.Body().SetAttributeValue("tls_cert_file", cty.StringVal(makeAbsolute(c.CertFile, "/etc/vault.d")))
			listenerBlock.Body().SetAttributeValue("tls_key_file", cty.StringVal(makeAbsolute(c.KeyFile, "/etc/vault.d")))
		} else {
			listenerBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
		}
	}

	return generate(f)
}
