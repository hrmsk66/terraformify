g terraformify

An experimental CLI that generates TF files to manage existing Fastly services with Terraform

https://user-images.githubusercontent.com/30490956/169726673-33ecccf7-ae40-4ebd-acf7-e4d457d4f510.mp4

## Installation

```
go install github.com/hrmsk66/terraformify@latest
```

## Configuration

terraformify requires read permissions to the target Fastly resource.
Choose one of the following options to give terraformify access to your API token:

- Include the token explicitly on each command you run using the `--api-key` or `-k` flags.
- Set a `FASTLY_API_KEY` environment variable.

## Usage

Run the command in an empty directory

```
mkdir test && cd test
terraformify service <service-id>
```

**Note:** The generated main.tf may contain sensitive information such as API keys for logging endpoints; before committing to Git, replace them with variables so that the configuration file does not contain such information.

### Interactive mode

By default, terraformify imports all resources associated with the service, such as ACL entries, dictionary items, WAF..etc. To interactively select which resources to import, use the `--interactive` or `-i` flag.

```
terraformify service <service-id> -i
```

### Import specific version

By default, either the active version will be imported, or the latest version if no version is active. Alternatively, a specific version of the service can be selected by passing version number to the `--version` or `-v` flag.

```
terraformify service <service-id> -v <version-number>
```

### Manage associated resources

By default, the `manage_*` attribute is not set so that these resources can be managed externally.

| Resource Name                          | Attribute Name                                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| fastly_service_acl_entries             | [manage_entries](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_acl_entries)              |
| fastly_service_dictionary_items        | [manage_items](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dictionary_items)           |
| fastly_service_dynamic_snippet_content | [manage_snippets](https://registry.terraform.io/providers/fastly/fastly/latest/docs/resources/service_dynamic_snippet_content) |

To set the attributes to true and manage the resource with Terraform, use the `--manage-all` or `-m` flag.

```
terraformify service <service-id> -m
```

## License

MIT License
