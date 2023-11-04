# terraformify

An experimental CLI that generates Terraform files for managing existing Fastly services.

> [!IMPORTANT]
> **Known Issue: Incompatibility with Terraform 1.4.6 and Later**<br>
> `terraformify` is not compatible with Terraform versions 1.4.6 and later because of changes in the Terraform binary. To avoid any issues, please use `terraformify` with Terraform version 1.4.5 or earlier. For more information, see the issue at https://github.com/hrmsk66/terraformify/issues/49.

## Installation / Upgrade

```
go install github.com/hrmsk66/terraformify@latest
```

Alternatively, download the prebuild binary from [the latest release](https://github.com/hrmsk66/terraformify/releases/latest).

## Configuration

`terraformify` requires read access to your Fastly resources. Choose one of the following options to give `terraformify` access to your API token:

- Include the token explicitly on each command you run using the `--api-key` or `-k` flags.
- Set a `FASTLY_API_KEY` environment variable.

## Usage

Run `terraformify` command in an empty directory or in an existing TF directory.

> [!IMPORTANT]
> Executing the command within a directory containing existing TF files will alter the current state file and may modify other files (notably `variables.tf` and `terraform.tfvars`). It's advisable to back up your TF files before importing a new service.

### Importing VCL Service

```
terraformify service vcl <service-id>
```

### Importing Compute@Edge Service

For Compute services, provide the service ID and the path to the WASM package as arguments:

```
terraformify service compute <service-id> <path-to-package>
```

For more detailed usage instructions, including available flags and commands, see the [Usage Documentation](docs/USAGE.md).

## Supported Resources

`terraformify` supports the import of both Compute and VCL services, along with their associated resources. The following resources are supported:

### Supported Resources for VCL

- [fastly_service_vcl](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_vcl)
- [fastly_service_acl_entries](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_acl_entries)
- [fastly_service_dictionary_items](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dictionary_items)
- [fastly_service_dynamic_snippet_content](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dynamic_snippet_content)
- [fastly_service_waf_configuration](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_waf_configuration)

### Supported Resources for Compute

- [fastly_service_compute](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_compute)
- [fastly_configstore](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/configstore)
- [fastly_configstore_entries](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/configstore_entries)
- [fastly_secretstore](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/secretstore)
- [fastly_kvstore](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/kvstore)
- [fastly_service_dictionary_items](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dictionary_items)

## License

MIT License
