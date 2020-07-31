package config

import (
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func NewConsulConfiguration(
	datacenter string,
	bindAddr string,
	advertiseAddr string,
	clientAddr string,
	server bool,
	bootstrapExpect int64,
	retryJoin []string,
	encrypt string,
	enableTLS bool) string {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.SetAttributeValue("datacenter", cty.StringVal(datacenter))
	rootBody.SetAttributeValue("data_dir", cty.StringVal("/opt/consul"))

	if len(bindAddr) != 0 {
		rootBody.SetAttributeValue("bind_addr", cty.StringVal(bindAddr))
	}

	if len(advertiseAddr) != 0 {
		rootBody.SetAttributeValue("advertise_addr", cty.StringVal(advertiseAddr))
	}

	if len(clientAddr) != 0 {
		rootBody.SetAttributeValue("client_addr", cty.StringVal(clientAddr))
	}

	if len(retryJoin) != 0 {
		rootBody.SetAttributeValue("retry_join", cty.ListVal(transform(retryJoin)))
	}

	if server {
		rootBody.SetAttributeValue("ui", cty.BoolVal(true))
		rootBody.SetAttributeValue("server", cty.BoolVal(true))
		rootBody.SetAttributeValue("bootstrap_expect", cty.NumberIntVal(bootstrapExpect))
	}

	if len(encrypt) != 0 {
		rootBody.SetAttributeValue("encrypt", cty.StringVal(encrypt))
	}

	if enableTLS {
		rootBody.SetAttributeValue("ca_file", cty.StringVal("/etc/consul.d/consul-agent-ca.pem"))
		rootBody.SetAttributeValue("cert_file", cty.StringVal("/etc/consul.d/consul-agent-cert.pem"))
		rootBody.SetAttributeValue("key_file", cty.StringVal("/etc/consul.d/consul-agent-key.pem"))
		rootBody.SetAttributeValue("verify_incoming", cty.BoolVal(true))
		rootBody.SetAttributeValue("verify_outgoing", cty.BoolVal(true))
		rootBody.SetAttributeValue("verify_server_hostname", cty.BoolVal(true))
	}

	return string(f.Bytes())
}

func NewNomadConfiguration(
	datacenter string,
	bindAddr string,
	advertiseAddr string,
	server bool,
	client bool,
	bootstrapExpect int64,
	retryJoin []string,
	encrypt string,
	enableTLS bool) string {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.SetAttributeValue("datacenter", cty.StringVal(datacenter))
	rootBody.SetAttributeValue("data_dir", cty.StringVal("/opt/nomad"))

	if len(bindAddr) != 0 {
		addressesBlock := rootBody.AppendNewBlock("addresses", []string{})
		addressesBlock.Body().SetAttributeValue("http", cty.StringVal(bindAddr))
		addressesBlock.Body().SetAttributeValue("rpc", cty.StringVal(bindAddr))
		addressesBlock.Body().SetAttributeValue("serf", cty.StringVal(bindAddr))
	}

	if len(advertiseAddr) != 0 {
		addressesBlock := rootBody.AppendNewBlock("advertise", []string{})
		addressesBlock.Body().SetAttributeValue("http", cty.StringVal(advertiseAddr))
		addressesBlock.Body().SetAttributeValue("rpc", cty.StringVal(advertiseAddr))
		addressesBlock.Body().SetAttributeValue("serf", cty.StringVal(advertiseAddr))
	}

	if server {
		serverBlock := rootBody.AppendNewBlock("server", []string{})
		serverBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))
		serverBlock.Body().SetAttributeValue("bootstrap_expect", cty.NumberIntVal(bootstrapExpect))

		if len(retryJoin) != 0 {
			serverJoinBlock := serverBlock.Body().AppendNewBlock("server_join", []string{})
			serverJoinBlock.Body().SetAttributeValue("retry_join", cty.ListVal(transform(retryJoin)))
		}

		if len(encrypt) != 0 {
			serverBlock.Body().SetAttributeValue("encrypt", cty.StringVal(encrypt))
		}
	}

	if client {
		clientBlock := rootBody.AppendNewBlock("client", []string{})
		clientBlock.Body().SetAttributeValue("enabled", cty.BoolVal(true))

		if len(retryJoin) != 0 {
			serverJoinBlock := clientBlock.Body().AppendNewBlock("server_join", []string{})
			serverJoinBlock.Body().SetAttributeValue("retry_join", cty.ListVal(transform(retryJoin)))
		}
	}

	if enableTLS {
		tlsBlock := rootBody.AppendNewBlock("tls", []string{})
		tlsBlock.Body().SetAttributeValue("http", cty.BoolVal(true))
		tlsBlock.Body().SetAttributeValue("rpc", cty.BoolVal(true))
		tlsBlock.Body().SetAttributeValue("ca_file", cty.StringVal("/etc/nomad.d/nomad-agent-ca.pem"))
		tlsBlock.Body().SetAttributeValue("cert_file", cty.StringVal("/etc/nomad.d/nomad-agent-cert.pem"))
		tlsBlock.Body().SetAttributeValue("key_file", cty.StringVal("/etc/nomad.d/nomad-agent-key.pem"))
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
