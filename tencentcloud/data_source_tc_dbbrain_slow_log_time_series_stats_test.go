package tencentcloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTencentCloudDbbrainSlowLogTimeSeriesStatsDataSource_basic(t *testing.T) {
	t.Parallel()
	loc, _ := time.LoadLocation("Asia/Chongqing")
	startTime := time.Now().AddDate(0, 0, -1).In(loc).Format("2006-01-02 15:04:05")
	endTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDbbrainSlowLogTimeSeriesStatsDataSource, defaultDbBrainInstanceId, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_dbbrain_slow_log_time_series_stats.test"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "period"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "time_series.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "time_series.0.count"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "time_series.0.timestamp"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "series_data.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "series_data.0.series.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "series_data.0.series.0.values.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_dbbrain_slow_log_time_series_stats.test", "series_data.0.timestamp.#"),
				),
			},
		},
	})
}

const testAccDbbrainSlowLogTimeSeriesStatsDataSource = `

data "tencentcloud_dbbrain_slow_log_time_series_stats" "test" {
  instance_id = "%s"
  start_time = "%s"
  end_time = "%s"
  product = "mysql"
}

`
