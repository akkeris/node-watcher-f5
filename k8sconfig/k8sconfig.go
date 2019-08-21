package k8sconfig

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	vault "github.com/akkeris/vault-client"
	"strings"
	"text/template"
)

const tokenconfigtemplate = `apiVersion: v1
clusters:
- cluster:
    server: https://{{ .Apiserverurl }}
  name: {{ .Cluster }}
contexts:
- context:
    cluster: {{ .Cluster }}
    user: {{ .Cluster }}
  name: {{ .Cluster }}
current-context: {{ .Cluster }}
kind: Config
preferences: {}
users:
- name: {{ .Cluster }}
  user:
    token: {{ .Token }}
`

const certconfigtemplate = `apiVersion: v1
clusters:
- cluster:
    certificate-authority: ca.pem
    server: https://{{ .Apiserverurl }}
  name: {{ .Cluster }}
contexts:
- context:
    cluster: {{ .Cluster }}
    user: {{ .Cluster }}
  name: {{ .Cluster }}
current-context: {{ .Cluster }}
kind: Config
preferences: {}
users:
- name: {{ .Cluster }}
  user:
    client-certificate: admin.pem
    client-key: admin-key.pem
`

type Config struct {
	Apiserverurl string
	Cluster      string
	Token        string
}

func CreateConfig() {
	var config Config

	if os.Getenv("KUBERNETES_CLIENT_TYPE") == "token" {
		config.Cluster = os.Getenv("CLUSTER")
		config.Apiserverurl = os.Getenv("KUBERNETES_API_SERVER")
		config.Token = getToken()

		it := template.Must(template.New("config").Parse(tokenconfigtemplate))
		var ib bytes.Buffer
		iwr := bufio.NewWriter(&ib)
		err := it.Execute(iwr, config)
		if err != nil {
			fmt.Println(err)
		}
		iwr.Flush()
		err = ioutil.WriteFile("config", ib.Bytes(), 0755)
		if err != nil {
			fmt.Println(err)
		}

	}

	if os.Getenv("KUBERNETES_CLIENT_TYPE") == "cert" {
		getCerts()
		config.Cluster = os.Getenv("CLUSTER")
		config.Apiserverurl = os.Getenv("KUBERNETES_API_SERVER")

		it := template.Must(template.New("config").Parse(certconfigtemplate))
		var ib bytes.Buffer
		iwr := bufio.NewWriter(&ib)
		err := it.Execute(iwr, config)
		if err != nil {
			fmt.Println(err)
		}
		iwr.Flush()
		err = ioutil.WriteFile("config", ib.Bytes(), 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func getToken() (t string) {
	tokensecret := os.Getenv("KUBERNETES_TOKEN_SECRET")
	token := vault.GetField(tokensecret, "token")
	return token
}

func getCerts() {
	//strings.Replace(vault.GetField(path, "key"), "\\\\n", "\n", -1)

	kubernetescertsecret := os.Getenv("KUBERNETES_CERT_SECRET")
	admincrt := strings.Replace(vault.GetField(kubernetescertsecret, "admin-crt"), "\\n", "\n", -1)
	adminkey := strings.Replace(vault.GetField(kubernetescertsecret, "admin-key"), "\\n", "\n", -1)
	cacrt := strings.Replace(vault.GetField(kubernetescertsecret, "ca-crt"), "\\n", "\n", -1)

	ca, err := os.Create("ca.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer ca.Close()

	_, err = ca.WriteString(cacrt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	crt, err := os.Create("admin.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer crt.Close()

	_, err = crt.WriteString(admincrt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	key, err := os.Create("admin-key.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer key.Close()

	_, err = key.WriteString(adminkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
