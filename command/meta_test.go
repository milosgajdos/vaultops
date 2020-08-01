package command

import (
	"flag"
	"os"
	"sort"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

var (
	// default test values
	metaTest    *Meta
	fAddr       = "http://127.0.0.1:8200"
	fCACert     = "./ca/ca.pem"
	fCAPath     = "./ca/"
	fClientCert = "./ca/client.crt"
	fClientKey  = "./ca/client.key"
	fInsecure   = false
)

func setup() {
	metaTest = &Meta{
		UI:             &cli.BasicUi{},
		flagAddress:    fAddr,
		flagCACert:     fCACert,
		flagCAPath:     fCAPath,
		flagClientCert: fClientCert,
		flagClientKey:  fClientKey,
		flagInsecure:   fInsecure,
	}
}

func teardown() {}

func TestMain(m *testing.M) {
	// set up tests
	setup()
	// run the tests
	retCode := m.Run()
	// clean up setup
	teardown()
	// call with result of m.Run()
	os.Exit(retCode)
}

// ripped off https://github.com/hashicorp/vault/blob/master/meta/meta_test.go
func TestFlagSet(t *testing.T) {
	cases := []struct {
		Flags    FlagSetFlags
		Expected []string
	}{
		{
			FlagSetNone,
			[]string{},
		},
		{
			FlagSetServer,
			[]string{"address", "ca-cert", "ca-path", "client-cert", "client-key", "tls-skip-verify", "redact", "key-store", "kms-provider", "aws-kms-id", "gcp-kms-crypto-key", "gcp-kms-key-ring", "gcp-kms-region", "gcp-kms-project", "storage-bucket", "storage-key", "key-local-path", "namespace"},
		},
	}

	for _, tc := range cases {
		var m Meta
		fs := m.FlagSet("foo", tc.Flags)

		actual := make([]string, 0, 0)
		fs.VisitAll(func(f *flag.Flag) {
			actual = append(actual, f.Name)
		})
		sort.Strings(actual)
		sort.Strings(tc.Expected)
		assert.EqualValues(t, tc.Expected, actual)
	}
}

func TestConfig(t *testing.T) {
	// empty address
	config, err := metaTest.Config("")
	assert.NotNil(t, config)
	assert.NoError(t, err)
	assert.Equal(t, metaTest.flagAddress, config.Address)
	// empty Meta.flagAddress
	addr := metaTest.flagAddress
	metaTest.flagAddress = ""
	config, err = metaTest.Config("")
	assert.NotNil(t, config)
	assert.NoError(t, err)
	metaTest.flagAddress = addr
	// pass in some address
	addr = "http://vault:8200"
	config, err = metaTest.Config(addr)
	assert.NotNil(t, config)
	assert.NoError(t, err)
	assert.Equal(t, addr, config.Address)
	// empty TLS config
	m := &Meta{}
	config, err = m.Config("")
	assert.NotNil(t, config)
	assert.NoError(t, err)
	// inject weird evn var
	os.Setenv("VAULT_SKIP_VERIFY", "foobar")
	defer os.Setenv("VAULT_SKIP_VERIFY", "")
	config, err = metaTest.Config("")
	assert.Nil(t, config)
	assert.Error(t, err)
}

func TestClient(t *testing.T) {
	client, err := metaTest.Client("", "")
	assert.NotNil(t, client)
	assert.NoError(t, err)
	// supply address
	addr := "http://vault:8200"
	token := "token"
	client, err = metaTest.Client(addr, token)
	assert.NotNil(t, client)
	assert.NoError(t, err)
	assert.Equal(t, token, client.Token())
	// mess up config = mess up client
	os.Setenv("VAULT_SKIP_VERIFY", "foobar")
	defer os.Setenv("VAULT_SKIP_VERIFY", "")
	client, err = metaTest.Client("", "")
	assert.Nil(t, client)
	assert.Error(t, err)
}
