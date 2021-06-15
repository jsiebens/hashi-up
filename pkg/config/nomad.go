package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type NomadConfig struct {
	Datacenter      string
	BindAddr        string
	AdvertiseAddr   string
	Server          bool
	Client          bool
	NodeClass       string
	BootstrapExpect int64
	RetryJoin       []string
	Encrypt         string
	CaFile          string
	CertFile        string
	KeyFile         string
	EnableACL       bool
}

func (c NomadConfig) EnableTLS() bool {
	return len(c.CaFile) != 0 && len(c.CertFile) != 0 && len(c.KeyFile) != 0
}

func (c NomadConfig) GenerateConfigFile() string {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.SetAttributeValue("datacenter", cty.StringVal(c.Datacenter))
	rootBody.SetAttributeValue("data_dir", cty.StringVal("/opt/nomad"))

	if len(c.BindAddr) != 0 {
		addressesBlock := rootBody.AppendNewBlock("addresses", []string{})
		addressesBlock.Body().SetAttributeValue("http", cty.StringVal(c.BindAddr))
		addressesBlock.Body().SetAttributeValue("rpc", cty.StringVal(c.BindAddr))
		addressesBlock.Body().SetAttributeValue("serf", cty.StringVal(c.BindAddr))
	}

	if len(c.AdvertiseAddr) != 0 {
		addressesBlock := rootBody.AppendNewBlock("advertise", []string{})
		addressesBlock.Body().SetAttributeValue("http", cty.StringVal(c.AdvertiseAddr))
		addressesBlock.Body().SetAttributeValue("rpc", cty.StringVal(c.AdvertiseAddr))
		addressesBlock.Body().SetAttributeValue("serf", cty.StringVal(c.AdvertiseAddr))
	}

	if c.Server {
		serverBlock := rootBody.AppendNewBlock("server", []string{})
		serverBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))
		if c.BootstrapExpect > 0 {
			serverBlock.Body().SetAttributeValue("bootstrap_expect", cty.NumberIntVal(c.BootstrapExpect))
		}

		if len(c.RetryJoin) != 0 {
			serverJoinBlock := serverBlock.Body().AppendNewBlock("server_join", []string{})
			serverJoinBlock.Body().SetAttributeValue("retry_join", cty.ListVal(transform(c.RetryJoin)))
		}

		if len(c.Encrypt) != 0 {
			serverBlock.Body().SetAttributeValue("encrypt", cty.StringVal(c.Encrypt))
		}
	}

	if c.Client {
		clientBlock := rootBody.AppendNewBlock("client", []string{})
		clientBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))

		if len(c.NodeClass) != 0 {
			clientBlock.Body().SetAttributeValue("node_class", cty.StringVal(c.NodeClass))
		}

		if len(c.RetryJoin) != 0 {
			serverJoinBlock := clientBlock.Body().AppendNewBlock("server_join", []string{})
			serverJoinBlock.Body().SetAttributeValue("retry_join", cty.ListVal(transform(c.RetryJoin)))
		}
	}

	if c.EnableTLS() {
		tlsBlock := rootBody.AppendNewBlock("tls", []string{})
		tlsBlock.Body().SetAttributeValue("http", cty.BoolVal(true))
		tlsBlock.Body().SetAttributeValue("rpc", cty.BoolVal(true))
		tlsBlock.Body().SetAttributeValue("ca_file", cty.StringVal(makeAbsolute(c.CaFile, "/etc/nomad.d")))
		tlsBlock.Body().SetAttributeValue("cert_file", cty.StringVal(makeAbsolute(c.CertFile, "/etc/nomad.d")))
		tlsBlock.Body().SetAttributeValue("key_file", cty.StringVal(makeAbsolute(c.KeyFile, "/etc/nomad.d")))
	}

	if c.EnableACL {
		aclBlock := rootBody.AppendNewBlock("acl", []string{})
		aclBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))
	}

	return generate(f)
}
