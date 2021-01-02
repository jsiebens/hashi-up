package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"path/filepath"
)

type ConsulConfig struct {
	Datacenter      string
	BindAddr        string
	AdvertiseAddr   string
	ClientAddr      string
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

	if c.Server {
		rootBody.SetAttributeValue("ui", cty.BoolVal(true))
		rootBody.SetAttributeValue("server", cty.BoolVal(true))
		rootBody.SetAttributeValue("bootstrap_expect", cty.NumberIntVal(c.BootstrapExpect))
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

type NomadConfig struct {
	Datacenter      string
	BindAddr        string
	AdvertiseAddr   string
	Server          bool
	Client          bool
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
		serverBlock.Body().SetAttributeValue("bootstrap_expect", cty.NumberIntVal(c.BootstrapExpect))

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

	return string(f.Bytes())
}

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
		rootBody.SetAttributeValue("cluster_addr", cty.StringVal(c.ConsulAddr))
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

	return string(f.Bytes())
}

func transform(vs []string) []cty.Value {
	vsm := make([]cty.Value, len(vs))
	for i, v := range vs {
		vsm[i] = cty.StringVal(v)
	}
	return vsm
}

func makeAbsolute(path string, base string) string {
	_, filename := filepath.Split(expandPath(path))
	return base + "/" + filename
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}
