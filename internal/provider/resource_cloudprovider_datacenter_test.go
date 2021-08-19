package provider

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/harness-io/harness-go-sdk/harness/api/cac"
	"github.com/harness-io/harness-go-sdk/harness/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDataCenterCloudProviderConnector(t *testing.T) {

	var (
		name         = fmt.Sprintf("%s_%s", t.Name(), utils.RandStringBytes(4))
		updatedName  = fmt.Sprintf("%s_updated", name)
		resourceName = "harness_cloudprovider_datacenter.test"
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCloudProviderDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataCenterCloudProvider(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccCheckDataCenterCloudProviderExists(t, resourceName, name),
				),
			},
			{
				Config: testAccResourceDataCenterCloudProvider(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCenterCloudProviderExists(t, resourceName, name),
				),
				ExpectError: regexp.MustCompile("name is immutable"),
			},
		},
	})
}

func testAccResourceDataCenterCloudProvider(name string) string {
	return fmt.Sprintf(`
		resource "harness_cloudprovider_datacenter" "test" {
			name = "%[1]s"
		}	
`, name)
}

func testAccCheckDataCenterCloudProviderExists(t *testing.T, resourceName, cloudProviderName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cp := &cac.PhysicalDatacenterCloudProvider{}
		err := testAccGetCloudProvider(resourceName, state, cp)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccGetCloudProvider(resourceName string, state *terraform.State, respObj interface{}) error {
	r := testAccGetResource(resourceName, state)
	if r == nil {
		return errors.New("Resource not found")
	}

	c := testAccGetApiClientFromProvider()
	name := r.Primary.Attributes["name"]

	err := c.ConfigAsCode().GetCloudProviderByName(name, respObj)
	if err != nil {
		return err
	}

	return nil
}

func testAccCloudProviderDestroy(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cp := &cac.PhysicalDatacenterCloudProvider{}
		err := testAccGetCloudProvider(resourceName, state, &cp)
		if err == nil {
			return errors.New("found cloud provider")
		}

		return nil
	}
}