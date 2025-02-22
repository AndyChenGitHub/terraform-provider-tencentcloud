package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// go test -i; go test -test.run TestAccTencentCloudRedisInstanceShardsDataSource_basic -v
func TestAccTencentCloudRedisInstanceShardsDataSource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedisInstanceShardsDataSource,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_redis_instance_shards.instance_shards"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.connected"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.keys"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.role"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.runid"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.shard_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.shard_name"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.slots"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.storage"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_redis_instance_shards.instance_shards", "instance_shards.0.storage_slope"),
				),
			},
		},
	})
}

const testAccRedisInstanceShardsDataSource = testAccRedisInstanceCluster + `

data "tencentcloud_redis_instance_shards" "instance_shards" {
	instance_id = tencentcloud_redis_instance.redis_cluster.id
	filter_slave = true
}

`
