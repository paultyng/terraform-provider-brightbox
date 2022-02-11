package brightbox

import (
	"fmt"
	"log"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validPermissionsGroups = []string{"full", "storage"}

func resourceBrightboxAPIClient() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Brightbox API Client resource",
		Create:      resourceBrightboxAPIClientCreate,
		Read:        resourceBrightboxAPIClientRead,
		Update:      resourceBrightboxAPIClientUpdate,
		Delete:      resourceBrightboxAPIClientDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"account": {
				Description: "The account the API client relates to",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"description": {
				Description: "Verbose Description of this client",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"name": {
				Description: "Human Readable Name",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"permissions_group": {
				Description:  "Summary of the permissions granted to the client (full, storage)",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      validPermissionsGroups[0],
				ValidateFunc: validation.StringInSlice(validPermissionsGroups, false),
			},

			"secret": {
				Description: "A shared secret the client must present when authenticating",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceBrightboxAPIClientCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Api Client")
	apiClientOpts := &brightbox.APIClientOptions{}
	err := addUpdateableAPIClientOptions(d, apiClientOpts)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Api Client create configuration: %#v", apiClientOpts)
	apiClient, err := client.CreateAPIClient(apiClientOpts)
	if err != nil {
		return fmt.Errorf("Error creating Api Client: %s", err)
	}

	d.SetId(apiClient.ID)

	return setAPIClientAttributes(d, apiClient)
}

func resourceBrightboxAPIClientRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	apiClient, err := client.APIClient(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Api Client details: %s", err)
	}
	if apiClient.RevokedAt != nil {
		log.Printf("[WARN] Api Client revoked, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Api Client read: %#v", apiClient)
	return setAPIClientAttributes(d, apiClient)
}

func resourceBrightboxAPIClientDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Deleting Api Client %s", d.Id())
	err := client.DestroyAPIClient(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Api Client (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxAPIClientUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	apiClientOpts := &brightbox.APIClientOptions{
		ID: d.Id(),
	}
	err := addUpdateableAPIClientOptions(d, apiClientOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Api Client update configuration: %#v", apiClientOpts)

	apiClient, err := client.UpdateAPIClient(apiClientOpts)
	if err != nil {
		return fmt.Errorf("Error updating Api Client (%s): %s", apiClientOpts.ID, err)
	}

	return setAPIClientAttributes(d, apiClient)
}

func addUpdateableAPIClientOptions(
	d *schema.ResourceData,
	opts *brightbox.APIClientOptions,
) error {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	assignString(d, &opts.PermissionsGroup, "permissions_group")
	return nil
}

func setAPIClientAttributes(
	d *schema.ResourceData,
	apiClient *brightbox.APIClient,
) error {
	d.Set("name", apiClient.Name)
	d.Set("description", apiClient.Description)
	d.Set("permissions_group", apiClient.PermissionsGroup)
	d.Set("account", apiClient.Account.ID)

	// Only update the secret if it is set
	if apiClient.Secret != "" {
		d.Set("secret", apiClient.Secret)
	}
	return nil
}
