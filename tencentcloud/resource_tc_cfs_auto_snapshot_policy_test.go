package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTencentCloudCfsAutoSnapshotPolicyResource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCfsAutoSnapshotPolicy,
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_cfs_auto_snapshot_policy.auto_snapshot_policy", "id")),
			},
			{
				ResourceName:      "tencentcloud_cfs_auto_snapshot_policy.auto_snapshot_policy",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccCfsAutoSnapshotPolicy = `

resource "tencentcloud_cfs_auto_snapshot_policy" "auto_snapshot_policy" {
  day_of_week = "1,2"
  hour = "2,3"
  policy_name = "policy_name"
  alive_days = 7
}

`
