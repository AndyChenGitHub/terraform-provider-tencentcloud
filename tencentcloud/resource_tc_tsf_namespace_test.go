package tencentcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// go test -i; go test -test.run TestAccTencentCloudTsfNamespaceResource_basic -v
func TestAccTencentCloudTsfNamespaceResource_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCommon(t, ACCOUNT_TYPE_TSF) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTsfNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTsfNamespace,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTsfNamespaceExists("tencentcloud_tsf_namespace.namespace"),
					resource.TestCheckResourceAttrSet("tencentcloud_tsf_namespace.namespace", "id"),
					resource.TestCheckResourceAttr("tencentcloud_tsf_namespace.namespace", "namespace_name", "terraform-namespace-name"),
					resource.TestCheckResourceAttr("tencentcloud_tsf_namespace.namespace", "namespace_desc", "terraform-test"),
					resource.TestCheckResourceAttr("tencentcloud_tsf_namespace.namespace", "namespace_type", "DEF"),
					resource.TestCheckResourceAttr("tencentcloud_tsf_namespace.namespace", "is_ha_enable", "0"),
				),
			},
			// {
			// 	ResourceName:      "tencentcloud_tsf_namespace.namespace",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func testAccCheckTsfNamespaceDestroy(s *terraform.State) error {
	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)
	service := TsfService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tencentcloud_tsf_namespace" {
			continue
		}

		res, err := service.DescribeTsfNamespaceById(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if res != nil {
			return fmt.Errorf("tsf namespace %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckTsfNamespaceExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		logId := getLogId(contextNil)
		ctx := context.WithValue(context.TODO(), logIdKey, logId)

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("resource %s is not found", r)
		}

		service := TsfService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}
		res, err := service.DescribeTsfNamespaceById(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if res == nil {
			return fmt.Errorf("tsf namespace %s is not found", rs.Primary.ID)
		}

		return nil
	}
}

const testAccTsfNamespace = `

resource "tencentcloud_tsf_namespace" "namespace" {
	namespace_name = "terraform-namespace-name"
	namespace_desc = "terraform-test"
	namespace_type = "DEF"
	is_ha_enable = "0"
}

`
