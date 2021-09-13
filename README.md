# TF Latest Version

Tool to make sure the latest `required_providers` and `helm_releases` are used.


## How To

To update all provider and helm versions in the current directory and its sub directories.
```sh
tf-latest-version --path .
```

Versions can be ignored, causing the updater to skip them, by adding a comment before the resource.
```hcl
terraform {
  required_version = "0.14.7"

  required_providers {
    #tf-latest-version:ignore
    helm = {
      source  = "hashicorp/helm"
      version = "2.1.1"
    }
  }
}

#tf-latest-version:ignore
resource "helm_release" "cert_manager" {
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  name       = "cert-manager"
  version    = "v1.3.1"
}
```

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

