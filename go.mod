module github.com/xenitab/tf-provider-latest

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/minamijoyo/tfupdate v0.5.1
	github.com/spf13/afero v1.6.0
	github.com/stretchr/testify v1.7.0
	github.com/zclconf/go-cty v1.9.0
	helm.sh/helm/v3 v3.6.2
	k8s.io/client-go v11.0.0+incompatible // indirect
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/go-macaron/cors => github.com/go-macaron/cors v0.0.0-20190925001837-b0274f40d4c7
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
