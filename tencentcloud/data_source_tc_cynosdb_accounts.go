/*
Use this data source to query detailed information of cynosdb accounts

Example Usage

```hcl
data "tencentcloud_cynosdb_accounts" "accounts" {
	cluster_id = "cynosdbmysql-bws8h88b"
	account_names = ["root"]
}
```
*/
package tencentcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	cynosdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cynosdb/v20190107"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func dataSourceTencentCloudCynosdbAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudCynosdbAccountsRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "The ID of cluster.",
			},

			"account_names": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of accounts to be filtered.",
			},

			"hosts": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of hosts to be filtered.",
			},

			"account_set": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Database account list.&amp;quot;&amp;quot;Note: This field may return null, indicating that no valid value can be obtained.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Account name of database.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The account description of database.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time.",
						},
						"update_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Update time.",
						},
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Host.",
						},
						"max_user_connections": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of user connections.",
						},
					},
				},
			},

			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceTencentCloudCynosdbAccountsRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_cynosdb_accounts.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	var clusterId string

	paramMap := make(map[string]interface{})
	if v, ok := d.GetOk("cluster_id"); ok {
		clusterId = v.(string)
	}

	if v, ok := d.GetOk("account_names"); ok {
		accountNamesSet := v.(*schema.Set).List()
		paramMap["account_names"] = accountNamesSet
	}

	if v, ok := d.GetOk("hosts"); ok {
		hostsSet := v.(*schema.Set).List()
		paramMap["Hosts"] = helper.InterfacesStringsPoint(hostsSet)
	}

	service := CynosdbService{client: meta.(*TencentCloudClient).apiV3Conn}

	var accountSet []*cynosdb.Account

	err := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		response, e := service.DescribeCynosdbAccountsByFilter(ctx, clusterId, paramMap)
		if e != nil {
			return retryError(e)
		}
		accountSet = response.AccountSet
		return nil
	})
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(accountSet))
	tmpList := make([]map[string]interface{}, 0, len(accountSet))
	for _, item := range accountSet {
		ids = append(ids, *item.AccountName)
		account := make(map[string]interface{})
		account["account_name"] = item.AccountName
		account["description"] = item.Description
		account["create_time"] = item.CreateTime
		account["update_time"] = item.UpdateTime
		account["host"] = item.Host
		account["max_user_connections"] = item.MaxUserConnections

		tmpList = append(tmpList, account)
	}
	d.SetId(helper.DataResourceIdsHash(ids))
	d.Set("account_set", tmpList)
	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := writeToFile(output.(string), tmpList); e != nil {
			return e
		}
	}
	return nil
}
