# Detailed Usage Guide for `terraformify`

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

### Customizing the Resource Name

By default, `service` is used as the TF resource name. To customize it, use the `--resource-name` or `-n` flag.

```
terraformify service (vcl|compute) <service-id> [<path-to-package>] -n <resource-name>
```

The generated files and directories will be named after the TF resource being imported. If multiple services are to be managed, resource names should be specified to distinguish them.

### Interactive Mode

By default, the tool imports all resources associated with the service, such as ACL entries, dictionary items, WAF..etc. To interactively select which resources to import, use the `--interactive` or `-i` flag. Interactive mode is only for VCL services and does not work for Compute services.

```
terraformify service vcl <service-id> -i
```

### Importing Specific Version

By default, either the active version will be imported, or the latest version if no version is active. Alternatively, a specific version of the service can be selected by passing version number to the `--version` or `-v` flag.

```
terraformify service (vcl|compute) <service-id> [<path-to-package>] -v <version-number>
```

### force_destroy

By default, `force_destroy` is set to `false`. To set them to `true` and allow Terraform to destroy resources, use the `--force-destroy` or `-f` flag.

```
terraformify service (vcl|compute) <service-id> [<path-to-package>] -f
```

### Manage Associated Resources

By default, the `manage_*` attribute is not set so that these resources can be managed externally.

| Resource Name                          | Attribute Name                                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| fastly_service_acl_entries             | [manage_entries](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_acl_entries)              |
| fastly_service_dictionary_items        | [manage_items](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dictionary_items)           |
| fastly_service_dynamic_snippet_content | [manage_snippets](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dynamic_snippet_content) |

To set the attributes to true and manage the resource with Terraform, use the `--manage-all` or `-m` flag.

```
terraformify service (vcl|compute) <service-id> [<path-to-package>] -m
```

### Skip Editing the State File

By default, the tool updates `terraform.tfstate` directly. To disable this behavior and leave the state file untouched, use the `--skip-edit-state` or `-s` flag.

> [!IMPORTANT]
> This flag is intended for debugging purposes. Without updating the state file, Terraform detects diffs and `terraform apply` may result in the destruction and re-creation of associated resources, such as ACL entries and Dictionary items.

```
terraformify service (vcl|compute) <service-id> [<path-to-package>] -s
```
