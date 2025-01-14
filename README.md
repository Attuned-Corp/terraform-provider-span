# Terraform Provider for Span

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) or [OpenTofu](https://opentofu.org/docs/intro/install/)
- [Go](https://go.dev/doc/install) 1.23.x (to build the provider plugin)
- [golangci-lint](https://golangci-lint.run/) and [gotestsum](https://github.com/gotestyourself/gotestsum) to contribute & develop



## Status & Stability

The provider is currently *experimental* and in rapid iteration mode.
API stability is not guaranteed before locked versioning and publish to the terraform registry.
The provider is not currently published, so usage is recommended only in local testing environments
via the development override method specified below



## Developing The Provider

To iterate on the provider with local development

1. Make sure that you have satisfied the software from the requirements section above.
2. Checkout the repository
3. Build & install the provider via `make install`
4. Add a local development override to your `~/.terraformrc` file

```
provider_installation {
   dev_overrides {
      "registry.terraform.io/attuned-corp/span" = "<absolute path to your GOBIN directory>"
   }
   direct {}
}
```

5. Cd to the `example/local-install` folder, adjust the provider configuration as required and execute `terraform plan`
