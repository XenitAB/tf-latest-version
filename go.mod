module github.com/xenitab/tf-provider-latest

go 1.15

require (
	github.com/hashicorp/hcl/v2 v2.8.2
	github.com/hashicorp/terraform-config-inspect v0.0.0-20210209133302-4fd17a0faac2
	github.com/minamijoyo/tfupdate v0.4.3
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.7.0
	github.com/zclconf/go-cty v1.2.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/client-go v11.0.0+incompatible // indirect
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/go-macaron/cors => github.com/go-macaron/cors v0.0.0-20190925001837-b0274f40d4c7
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
