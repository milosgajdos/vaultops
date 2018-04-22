package command

import (
	"io/ioutil"
)

// makeTestFile creates a temporary test file and writes data into it
// It returns full path to newly created path  or error if the file fails to be created
func makeTestFile(data []byte) (string, error) {
	// create temp file for testing
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		return "", err
	}
	// write data to temp file
	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}

//func Test_getVaultBackends(t *testing.T) {
//	vaultRoles := []struct {
//		BEName           string
//		Name             string
//		AllowedDomains   string
//		AllowBareDomains bool
//		AllowAnyName     bool
//		EnforceHostnames bool
//		Organization     string
//	}{
//		{"k8s-1", "api", "kubernetes", true, true, false, "org1"},
//	}
//
//	vaultCerts := []struct {
//		Name       string
//		Type       string
//		Root       bool
//		Role       string
//		Store      bool
//		CommonName string
//		TTL        string
//		IPSans     string
//	}{
//		{"k8s-1", "internal", true, "", false, "comm-name", "100h", ""},
//		{"k8s-2", "", false, "foorole", false, "foo-name", "87h", "10.1.1.1"},
//	}
//	vaultBackends := []string{"k8s-1", "k8s-2"}
//
//	// test data
//	data := `backends:
//  - name: "` + vaultBackends[0] + `"
//    roles:
//      - name: "` + vaultRoles[0].Name + `"
//        allowed_domains: "` + vaultRoles[0].AllowedDomains + `"
//        allow_bare_domains: ` + fmt.Sprintf("%t", vaultRoles[0].AllowBareDomains) + `
//        allow_any_name: ` + fmt.Sprintf("%t", vaultRoles[0].AllowAnyName) + `
//        enforce_hostnames: ` + fmt.Sprintf("%t", vaultRoles[0].EnforceHostnames) + `
//        organization: "` + vaultRoles[0].Organization + `"
//    certificates:
//      - name: "` + vaultCerts[0].Name + `"
//        root: ` + fmt.Sprintf("%t", vaultCerts[0].Root) + `
//        common_name: "` + vaultCerts[0].CommonName + `"
//        ttl: "` + vaultCerts[0].TTL + `"
//        type: "` + vaultCerts[0].Type + `"
//  - name: "` + vaultBackends[1] + `"
//    certificates:
//      - name: "` + vaultCerts[1].Name + `"
//        common_name: "` + vaultCerts[1].CommonName + `"
//        ttl: "` + vaultCerts[1].TTL + `"
//        ip_sans: "` + vaultCerts[1].IPSans + `"
//        role: "` + vaultCerts[1].Role + `"
//        store: ` + fmt.Sprintf("%t", vaultCerts[1].Store) + `
//`
//	f, err := makeTestFile([]byte(data))
//	defer os.Remove(f)
//	assert.NoError(t, err)
//	backends, err := getVaultBackends(f)
//	assert.NoError(t, err)
//	assert.NotNil(t, backends)
//
//	for i := 0; i < len(backends); i++ {
//		assert.Equal(t, vaultBackends[i], backends[i].Name)
//		assert.Equal(t, vaultBackends[i], backends[i].Certs[0].Backend)
//	}
//
//	// random file path causes error
//	backends, err = getVaultBackends("foobar/dfd")
//	assert.Error(t, err)
//	// no parsed backends returns error
//	data = `foo:
//  - bar: foobar
//`
//	f2, err := makeTestFile([]byte(data))
//	defer os.Remove(f2)
//	assert.NoError(t, err)
//	backends, err = getVaultBackends(f2)
//	assert.Error(t, err)
//}
