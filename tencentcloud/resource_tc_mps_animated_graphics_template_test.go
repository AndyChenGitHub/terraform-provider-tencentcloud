package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTencentCloudMpsAnimatedGraphicsTemplateResource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMpsAnimatedGraphicsTemplate,
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_mps_animated_graphics_template.animated_graphics_template", "id")),
			},
			{
				Config: testAccMpsAnimatedGraphicsTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tencentcloud_mps_animated_graphics_template.animated_graphics_template", "id"),
					resource.TestCheckResourceAttr("tencentcloud_mps_animated_graphics_template.animated_graphics_template", "name", "terraform-for-test"),
				),
			},
			{
				ResourceName:      "tencentcloud_mps_animated_graphics_template.animated_graphics_template",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccMpsAnimatedGraphicsTemplate = `

resource "tencentcloud_mps_animated_graphics_template" "animated_graphics_template" {
  format              = "gif"
  fps                 = 20
  height              = 130
  name                = "terraform-test"
  quality             = 75
  resolution_adaptive = "open"
  width               = 140
}

`

const testAccMpsAnimatedGraphicsTemplateUpdate = `

resource "tencentcloud_mps_animated_graphics_template" "animated_graphics_template" {
  format              = "gif"
  fps                 = 20
  height              = 130
  name                = "terraform-for-test"
  quality             = 75
  resolution_adaptive = "open"
  width               = 140
}

`
