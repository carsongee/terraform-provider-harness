package provider

import (
	"context"
	"errors"

	"github.com/harness-io/harness-go-sdk/harness/api"
	"github.com/harness-io/harness-go-sdk/harness/api/cac"
	"github.com/harness-io/terraform-provider-harness/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudProviderTanzu() *schema.Resource {

	providerSchema := map[string]*schema.Schema{
		"endpoint": {
			Description: "The url of the Tanzu platform.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"skip_validation": {
			Description: "Skip validation of Tanzu configuration.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"username": {
			Description:   "The username to use to authenticate to Tanzu.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"username_secret_name"},
		},
		"username_secret_name": {
			Description:   "The name of the Harness secret containing the username to authenticate to Tanzu with.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"username"},
		},
		"password_secret_name": {
			Description: "The name of the Harness secret containing the password to use to authenticate to Tanzu.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}

	helpers.MergeSchemas(commonCloudProviderSchema(), providerSchema)

	return &schema.Resource{
		Description:   "Resource for creating a Tanzu cloud provider",
		CreateContext: resourceCloudProviderTanzuCreate,
		ReadContext:   resourceCloudProviderTanzuRead,
		UpdateContext: resourceCloudProviderTanzuUpdate,
		DeleteContext: resourceCloudProviderTanzuDelete,

		Schema: providerSchema,
	}
}

func resourceCloudProviderTanzuRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*api.Client)

	name := d.Get("name").(string)

	cp := &cac.PcfCloudProvider{}
	err := c.ConfigAsCode().GetCloudProviderByName(name, cp)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cp.Id)
	d.Set("name", cp.Name)
	d.Set("endpoint", cp.EndpointUrl)
	d.Set("skip_validation", cp.SkipValidation)
	d.Set("username", cp.Username)

	if cp.UsernameSecretId != nil {
		d.Set("username_secret_name", cp.UsernameSecretId.Name)
	}

	d.Set("password_secret_name", cp.Password.Name)

	scope, err := flattenUsageRestrictions(c, cp.UsageRestrictions)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("usage_scope", scope)

	return nil
}

func resourceCloudProviderTanzuCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*api.Client)

	input := cac.NewEntity(cac.ObjectTypes.PcfCloudProvider).(*cac.PcfCloudProvider)
	input.Name = d.Get("name").(string)
	input.EndpointUrl = d.Get("endpoint").(string)
	input.SkipValidation = d.Get("skip_validation").(bool)
	input.Username = d.Get("username").(string)

	if attr := d.Get("username_secret_name").(string); attr != "" {
		input.UsernameSecretId = &cac.SecretRef{
			Name: attr,
		}
	}

	input.Password = &cac.SecretRef{
		Name: d.Get("password_secret_name").(string),
	}

	restrictions, err := expandUsageRestrictions(c, d.Get("usage_scope").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}
	input.UsageRestrictions = restrictions

	cp, err := c.ConfigAsCode().UpsertPcfCloudProvider(input)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cp.Id)

	return nil
}

func resourceCloudProviderTanzuUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*api.Client)

	if d.HasChange("name") {
		return diag.FromErr(errors.New("name is immutable"))
	}

	cp := cac.NewEntity(cac.ObjectTypes.AzureCloudProvider).(*cac.PcfCloudProvider)
	cp.Name = d.Get("name").(string)
	cp.EndpointUrl = d.Get("endpoint").(string)
	cp.SkipValidation = d.Get("skip_validation").(bool)
	cp.Username = d.Get("username").(string)

	if attr := d.Get("username_secret_name").(string); attr != "" {
		cp.UsernameSecretId = &cac.SecretRef{
			Name: attr,
		}
	}

	cp.Password = &cac.SecretRef{
		Name: d.Get("password_secret_name").(string),
	}

	usageRestrictions, err := expandUsageRestrictions(c, d.Get("usage_scope").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}
	cp.UsageRestrictions = usageRestrictions

	_, err = c.ConfigAsCode().UpsertPcfCloudProvider(cp)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCloudProviderTanzuDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*api.Client)

	id := d.Get("id").(string)
	err := c.CloudProviders().DeleteCloudProvider(id)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}