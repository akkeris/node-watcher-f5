package k8sconfig

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

const tokenconfigtemplate = `apiVersion: v1
clusters:
- cluster:
    server: {{ .Apiserverurl }}
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
    server: {{ .Apiserverurl }}
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

}

func getToken() (t string) {
	return os.Getenv("KUBERNETES_TOKEN")
}

