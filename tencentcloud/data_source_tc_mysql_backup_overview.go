/*
Use this data source to query detailed information of mysql backup_overview

Example Usage

```hcl
data "tencentcloud_mysql_backup_overview" "backup_overview" {
  product = "mysql"
}
```
*/
package tencentcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func dataSourceTencentCloudMysqlBackupOverview() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudMysqlBackupOverviewRead,
		Schema: map[string]*schema.Schema{
			"product": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "The type of cloud database product to be queried, currently only supports `mysql`.",
			},

			"backup_count": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The total number of user backups in the current region (including data backups and log backups).",
			},

			"backup_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The total backup capacity of the user in the current region.",
			},

			"billing_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The billable capacity of the user&amp;#39;s backup in the current region, that is, the part that exceeds the gifted capacity.",
			},

			"free_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The free backup capacity obtained by the user in the current region.",
			},

			"remote_backup_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "The total capacity of off-site backup of the user in the current region. Note: This field may return null, indicating that no valid value can be obtained.",
			},

			"backup_archive_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Archive backup capacity, including data backup and log backup. Note: This field may return null, indicating that no valid value can be obtained.",
			},

			"backup_standby_volume": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Standard storage backup capacity, including data backup and log backup. Note: This field may return null, indicating that no valid value can be obtained.",
			},

			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceTencentCloudMysqlBackupOverviewRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_mysql_backup_overview.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	product := ""
	paramMap := make(map[string]interface{})
	if v, ok := d.GetOk("product"); ok {
		product = v.(string)
		paramMap["Product"] = helper.String(v.(string))
	}

	var backupCount *cdb.DescribeBackupOverviewResponseParams
	service := MysqlService{client: meta.(*TencentCloudClient).apiV3Conn}
	err := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		result, e := service.DescribeMysqlBackupOverviewByFilter(ctx, paramMap)
		if e != nil {
			return retryError(e)
		}
		backupCount = result
		return nil
	})
	if err != nil {
		return err
	}

	if backupCount.BackupCount != nil {
		_ = d.Set("backup_count", backupCount.BackupCount)
	}

	if backupCount.BackupVolume != nil {
		_ = d.Set("backup_volume", backupCount.BackupVolume)
	}

	if backupCount.BillingVolume != nil {
		_ = d.Set("billing_volume", backupCount.BillingVolume)
	}

	if backupCount.FreeVolume != nil {
		_ = d.Set("free_volume", backupCount.FreeVolume)
	}

	if backupCount.RemoteBackupVolume != nil {
		_ = d.Set("remote_backup_volume", backupCount.RemoteBackupVolume)
	}

	if backupCount.BackupArchiveVolume != nil {
		_ = d.Set("backup_archive_volume", backupCount.BackupArchiveVolume)
	}

	if backupCount.BackupStandbyVolume != nil {
		_ = d.Set("backup_standby_volume", backupCount.BackupStandbyVolume)
	}

	d.SetId(helper.DataResourceIdsHash([]string{product}))
	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := writeToFile(output.(string), backupCount); e != nil {
			return e
		}
	}
	return nil
}
