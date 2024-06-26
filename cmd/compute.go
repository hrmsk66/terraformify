package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/hrmsk66/terraformify/pkg/cli"
	"github.com/hrmsk66/terraformify/pkg/file"
	"github.com/hrmsk66/terraformify/pkg/prop"
	"github.com/hrmsk66/terraformify/pkg/terraform"
	"github.com/hrmsk66/terraformify/pkg/tfconf"
	"github.com/hrmsk66/terraformify/pkg/tfstate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// computeCmd represents the service command
var computeCmd = &cobra.Command{
	Use:          "compute <service-id>",
	Short:        "Generate TF files for an existing Fastly Compute service",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := cli.CreateLogFilter()
		log.Printf("[INFO] CLI version: %s", getVersion())
		log.SetOutput(filter)

		packagePath, err := cmd.Flags().GetString("package")
		if err != nil {
			return err
		}

		if packagePath != "" {
			if err = file.CheckPackage(packagePath); err != nil {
				return err
			}
		}

		workingDir, err := cmd.Flags().GetString("working-dir")
		if err != nil {
			return err
		}

		autoYes, err := cmd.Flags().GetBool("yes")
		if err != nil {
			return err
		}

		if err = file.CheckDir(workingDir, autoYes); err != nil {
			return err
		}

		resourceName, err := cmd.Flags().GetString("resource-name")
		if err != nil {
			return err
		}

		if err = file.CheckFile(workingDir, resourceName); err != nil {
			return err
		}

		apiKey := viper.GetString("api-key")
		if err = os.Setenv("FASTLY_API_KEY", apiKey); err != nil {
			return err
		}

		version, err := cmd.Flags().GetInt("version")
		if err != nil {
			return err
		}

		manageAll, err := cmd.Flags().GetBool("manage-all")
		if err != nil {
			return err
		}

		forceDestroy, err := cmd.Flags().GetBool("force-destroy")
		if err != nil {
			return err
		}

		skipEditState, err := cmd.Flags().GetBool("skip-edit-state")
		if err != nil {
			return err
		}

		testMode, err := cmd.Flags().GetBool("test-mode")
		if err != nil {
			return err
		}

		replaceDictionary, err := cmd.Flags().GetBool("replace-edge-dictionary")
		if err != nil {
			return err
		}

		c := cli.Config{
			ID:                args[0],
			Package:           packagePath,
			ResourceName:      resourceName,
			Version:           version,
			Directory:         workingDir,
			ManageAll:         manageAll,
			ForceDestroy:      forceDestroy,
			SkipEditState:     skipEditState,
			TestMode:          testMode,
			ReplaceDictionary: replaceDictionary,
		}

		return ImportCompute(c)
	},
}

func init() {
	serviceCmd.AddCommand(computeCmd)

	// Persistent flags
	serviceCmd.PersistentFlags().StringP("package", "p", "", "Path to the Compute service package file")
	serviceCmd.PersistentFlags().BoolP("replace-edge-dictionary", "r", false, "Generate TF files to replace edge dictionaries with config stores")
	serviceCmd.PersistentFlags().Lookup("replace-edge-dictionary").Hidden = true
}

func ImportCompute(c cli.Config) error {
	log.Printf("[INFO] Initializing Terraform")
	// Find Terraform binary
	tf, err := terraform.FindExec(c.Directory)
	if err != nil {
		return err
	}

	// Run "terraform version"
	if err = terraform.Version(tf); err != nil {
		return err
	}

	// Create provider.tf
	// Create temp*.tf with empty service resource blocks
	log.Printf("[INFO] Creating provider.tf and temp*.tf")
	tempf, err := file.CreateInitTerraformFiles(c.Directory)
	if err != nil {
		return err
	}

	// Run "terraform init"
	log.Printf(`[INFO] Running "terraform init"`)
	if err = terraform.Init(tf); err != nil {
		return err
	}

	// Create ComputeServiceResourceProp struct
	serviceProp := prop.NewComputeServiceResource(c.ID, c.ResourceName, c.Version)

	if err = terraform.Import(tf, serviceProp, tempf); err != nil {
		return err
	}

	// Get the config represented in HCL from the "terraform show" output
	log.Print(`[INFO] Running "terraform show" to get the current Terraform state in HCL format`)
	rawHCL, err := terraform.Show(tf)
	if err != nil {
		return err
	}

	// Parse HCL and obtain Terraform block props as a list of struct
	// to get the overall picture of the service configuration
	log.Print("[INFO] Parsing the HCL")
	hcl, err := tfconf.Load(rawHCL)
	if err != nil {
		return err
	}

	props, err := hcl.ParseServiceResource(serviceProp, &c)
	if err != nil {
		return err
	}

	// Iterate over the list of props and run terraform import for Dictionary items
	for _, p := range props {
		switch p := p.(type) {
		case *prop.DictionaryResource:
			if err = terraform.Import(tf, p, tempf); err != nil {
				return err
			}
		case *prop.LinkedResource:
			if c.TestMode {
				if err = terraform.RecursiveImport(tf, p, tempf); err != nil {
					return err
				}
			} else {
				t := cli.AskDataStoreType(p.GetName())
				p.SetDataStoreType(t)

				if err = terraform.Import(tf, p, tempf); err != nil {
					return err
				}
			}

			var entries *prop.LinkedResource
			entries, err = p.CloneForEntriesImport()
			if err == nil {
				if err = terraform.Import(tf, entries, tempf); err != nil {
					return err
				}
			}
		}
	}

	// temp*.tf no longer needed
	if err = tempf.Close(); err != nil {
		return err
	}
	if err = os.Remove(tempf.Name()); err != nil {
		return err
	}

	// Get the config represented in HCL from the "terraform show" output
	log.Print(`[INFO] Running "terraform show" to get the current Terraform state in HCL format`)
	rawHCL, err = terraform.Show(tf)
	if err != nil {
		return err
	}

	// Make changes to the configuration
	// log.Print("[INFO] Parsing the HCL and making corrections removing read-only attrs and replacing embedded VCL/logformat with the file function")
	log.Print("[INFO] Parsing the HCL and making corrections")
	hcl, err = tfconf.Load(rawHCL)
	if err != nil {
		return err
	}

	sensitiveAttrs, err := hcl.RewriteResources(serviceProp, props, &c)
	if err != nil {
		return err
	}

	if err := file.WriteTF(c.Directory, c.ResourceName, hcl.Bytes()); err != nil {
		return err
	}

	if err := file.WriteGitIgnore(c.Directory); err != nil {
		return err
	}

	if len(sensitiveAttrs) > 0 {
		variables := tfconf.BuildVariableDefinitions(sensitiveAttrs)
		if err := file.WriteVariablesTF(c.Directory, variables); err != nil {
			return err
		}

		tfvars := tfconf.BuildTFVars(sensitiveAttrs)
		if err := file.WriteTFVars(c.Directory, tfvars); err != nil {
			return err
		}
	}

	if c.SkipEditState {
		cli.BoldYellow(os.Stderr, "skip-edit-state flag detected. Leaving terraform.tfstate untouched")
	} else {
		log.Print(`[INFO] Setting "activate" in terraform.tfstate`)
		curState, err := tfstate.Load(c.Directory)
		if err != nil {
			return err
		}

		newState, err := curState.SetActivateAttribute(tfstate.SetActivateTemplateParams{
			ServiceId: c.ID,
		})
		if err != nil {
			return err
		}

		if c.Package != "" {
			log.Printf(`[INFO] Inserting "filename: %s" in terraform.tfstate`, c.Package)
			newState, err = newState.SetPackageFilename(tfstate.SetPackageFilenameParams{
				ServiceId:       c.ID,
				PackageFilename: c.Package,
			})
			if err != nil {
				return err
			}
		}

		if c.ManageAll {
			log.Print(`[INFO] Setting "manage_*" in terraform.tfstate`)
			newState, err = newState.SetManageAttributes(c.ID)
			if err != nil {
				return err
			}
		}

		if c.ForceDestroy {
			log.Print(`[INFO] Setting "force_destroy" in terraform.tfstate`)
			newState, err = newState.SetForceDestroy(tfstate.SetForceDestroyParams{
				ServiceId:    c.ID,
				ResourceType: serviceProp.GetType(),
			})
			if err != nil {
				return err
			}
		}

		for _, p := range props {
			switch p := p.(type) {
			case *prop.DictionaryResource:
				log.Printf(`[INFO] Inserting "index_key" in terraform.tfstate for %s`, p.GetRef())
				newState, err = newState.SetIndexKey(tfstate.SetIndexKeyParams{
					ServiceId:    c.ID,
					ResourceType: p.GetType(),
					ResourceName: p.GetNormalizedName(),
					Name:         p.GetName(),
				})
				if err != nil {
					return err
				}
			}
		}

		if len(sensitiveAttrs) > 0 {
			log.Print(`[INFO] Inserting items in "sensitive_attributes" in terraform.tfstate`)
			// Need to set once for each sensitive attribute
			blockTypes := map[string]struct{}{}
			for _, attr := range sensitiveAttrs {
				blockTypes[attr.BlockType] = struct{}{}
			}
			newState, err = newState.SetSensitiveAttributes(c.ID, blockTypes)
			if err != nil {
				return err
			}
		}

		if err = file.WriteTFState(c.Directory, newState.Bytes()); err != nil {
			return err
		}

		log.Print(`[INFO] Running "terraform refresh" to format the state file and check errors`)
		if err = terraform.Refresh(tf); err != nil {
			return err
		}
	}

	fmt.Fprintln(os.Stderr)
	cli.BoldGreen(os.Stderr, "Completed!")

	return nil
}
