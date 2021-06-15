package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type ConsulConfig struct {
	Datacenter      string
	BindAddr        string
	AdvertiseAddr   string
	ClientAddr      string
	DnsAddr         string
	HttpAddr        string
	HttpsAddr       string
	GrpcAddr        string
	Server          bool
	BootstrapExpect int64
	RetryJoin       []string
	Encrypt         string
	CaFile          string
	CertFile        string
	KeyFile         string
	AutoEncrypt     bool
	EnableACL       bool
	AgentToken      string
	EnableConnect   bool
	HttpsOnly       bool
}

func (c ConsulConfig) EnableTLS() bool {
	return c.AutoEncrypt || (len(c.CaFile) != 0 && len(c.CertFile) != 0 && len(c.KeyFile) != 0)
}

func (c ConsulConfig) GenerateConfigFile() string {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.SetAttributeValue("datacenter", cty.StringVal(c.Datacenter))
	rootBody.SetAttributeValue("data_dir", cty.StringVal("/opt/consul"))

	if len(c.BindAddr) != 0 {
		rootBody.SetAttributeValue("bind_addr", cty.StringVal(c.BindAddr))
	}

	if len(c.AdvertiseAddr) != 0 {
		rootBody.SetAttributeValue("advertise_addr", cty.StringVal(c.AdvertiseAddr))
	}

	if len(c.ClientAddr) != 0 {
		rootBody.SetAttributeValue("client_addr", cty.StringVal(c.ClientAddr))
	}

	if len(c.RetryJoin) != 0 {
		rootBody.SetAttributeValue("retry_join", cty.ListVal(transform(c.RetryJoin)))
	}

	portsBlock := rootBody.AppendNewBlock("ports", []string{})

	if c.EnableConnect {
		portsBlock.Body().SetAttributeValue("grpc", cty.NumberUIntVal(8502))
	}

	if c.EnableTLS() {
		portsBlock.Body().SetAttributeValue("https", cty.NumberUIntVal(8501))
		if c.HttpsOnly {
			portsBlock.Body().SetAttributeValue("http", cty.NumberIntVal(-1))
		}
	}

	addressesBlock := rootBody.AppendNewBlock("addresses", []string{})
	if len(c.DnsAddr) != 0 {
		addressesBlock.Body().SetAttributeValue("dns", cty.StringVal(c.DnsAddr))
	}
	if len(c.HttpAddr) != 0 {
		addressesBlock.Body().SetAttributeValue("http", cty.StringVal(c.HttpAddr))
	}
	if len(c.HttpsAddr) != 0 {
		addressesBlock.Body().SetAttributeValue("https", cty.StringVal(c.HttpsAddr))
	}
	if len(c.GrpcAddr) != 0 {
		addressesBlock.Body().SetAttributeValue("grpc", cty.StringVal(c.GrpcAddr))
	}

	if c.Server {
		rootBody.SetAttributeValue("ui", cty.BoolVal(true))
		rootBody.SetAttributeValue("server", cty.BoolVal(true))
		if c.BootstrapExpect > 0 {
			rootBody.SetAttributeValue("bootstrap_expect", cty.NumberIntVal(c.BootstrapExpect))
		}
	}

	if len(c.Encrypt) != 0 {
		rootBody.SetAttributeValue("encrypt", cty.StringVal(c.Encrypt))
	}

	if c.EnableTLS() {
		rootBody.SetAttributeValue("ca_file", cty.StringVal(makeAbsolute(c.CaFile, "/etc/consul.d")))

		if c.Server || !c.AutoEncrypt {
			rootBody.SetAttributeValue("cert_file", cty.StringVal(makeAbsolute(c.CertFile, "/etc/consul.d")))
			rootBody.SetAttributeValue("key_file", cty.StringVal(makeAbsolute(c.KeyFile, "/etc/consul.d")))
		}

		rootBody.SetAttributeValue("verify_incoming_rpc", cty.BoolVal(true))
		rootBody.SetAttributeValue("verify_outgoing", cty.BoolVal(true))
		rootBody.SetAttributeValue("verify_server_hostname", cty.BoolVal(true))

		if c.AutoEncrypt {
			autoTLSBlock := rootBody.AppendNewBlock("auto_encrypt", []string{})

			if c.Server {
				autoTLSBlock.Body().SetAttributeValue("allow_tls", cty.BoolVal(true))
			} else {
				rootBody.SetAttributeValue("verify_incoming_rpc", cty.BoolVal(false))
				autoTLSBlock.Body().SetAttributeValue("tls", cty.BoolVal(true))
			}
		}
	}

	if c.EnableACL {
		aclBlock := rootBody.AppendNewBlock("acl", []string{})
		aclBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))
		aclBlock.Body().SetAttributeValue("default_policy", cty.StringVal("deny"))
		aclBlock.Body().SetAttributeValue("down_policy", cty.StringVal("extend-cache"))
		aclBlock.Body().SetAttributeValue("enable_token_persistence", cty.BoolVal(true))

		if len(c.AgentToken) != 0 {
			tokensBlock := aclBlock.Body().AppendNewBlock("tokens", []string{})
			tokensBlock.Body().SetAttributeValue("agent", cty.StringVal(c.AgentToken))
		}
	}

	if c.EnableConnect {
		connectBlock := rootBody.AppendNewBlock("connect", []string{})
		connectBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))
	}

	return string(f.Bytes())
}
