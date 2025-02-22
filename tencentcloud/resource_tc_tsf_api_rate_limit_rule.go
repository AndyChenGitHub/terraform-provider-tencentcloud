/*
Provides a resource to create a tsf api_rate_limit_rule

Example Usage

```hcl
resource "tencentcloud_tsf_api_rate_limit_rule" "api_rate_limit_rule" {
  api_id = "api-xxxxxx"
  max_qps = 10
  usable_status = "enable"
}
```

Import

tsf api_rate_limit_rule can be imported using the id, e.g.

```
terraform import tencentcloud_tsf_api_rate_limit_rule.api_rate_limit_rule api_rate_limit_rule_id
```
*/
package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	tsf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tsf/v20180326"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudTsfApiRateLimitRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudTsfApiRateLimitRuleCreate,
		Read:   resourceTencentCloudTsfApiRateLimitRuleRead,
		Update: resourceTencentCloudTsfApiRateLimitRuleUpdate,
		Delete: resourceTencentCloudTsfApiRateLimitRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"api_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Api Id.",
			},

			"max_qps": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "qps value.",
			},

			"usable_status": {
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validateAllowedStringValue([]string{"enabled", "disabled"}),
				Description:  "Enabled/disabled, enabled/disabled, if not passed, it is enabled by default.",
			},

			"rule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "rule Id.",
			},

			"rule_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current limit name.",
			},

			"rule_content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Rule content.",
			},

			"tsf_rule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tsf Rule ID.",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "describe.",
			},

			"created_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "creation time.",
			},

			"updated_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "update time.",
			},
		},
	}
}

func resourceTencentCloudTsfApiRateLimitRuleCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_tsf_api_rate_limit_rule.create")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	var (
		request  = tsf.NewCreateApiRateLimitRuleWithDetailRespRequest()
		response = tsf.NewCreateApiRateLimitRuleWithDetailRespResponse()
		apiId    string
		ruleId   string
	)
	if v, ok := d.GetOk("api_id"); ok {
		apiId = v.(string)
		request.ApiId = helper.String(v.(string))
	}

	if v, _ := d.GetOk("max_qps"); v != nil {
		request.MaxQps = helper.IntUint64(v.(int))
	}

	if v, ok := d.GetOk("usable_status"); ok {
		request.UsableStatus = helper.String(v.(string))
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseTsfClient().CreateApiRateLimitRuleWithDetailResp(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s create tsf apiRateLimitRule failed, reason:%+v", logId, err)
		return err
	}

	ruleId = *response.Response.Result.RuleId
	d.SetId(apiId + FILED_SP + ruleId)

	return resourceTencentCloudTsfApiRateLimitRuleRead(d, meta)
}

func resourceTencentCloudTsfApiRateLimitRuleRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_tsf_api_rate_limit_rule.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := TsfService{client: meta.(*TencentCloudClient).apiV3Conn}

	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	apiId := idSplit[0]
	ruleId := idSplit[1]

	apiRateLimitRule, err := service.DescribeTsfApiRateLimitRuleById(ctx, apiId, ruleId)
	if err != nil {
		return err
	}

	if apiRateLimitRule == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `TsfApiRateLimitRule` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if apiRateLimitRule.ApiId != nil {
		_ = d.Set("api_id", apiRateLimitRule.ApiId)
	}

	if apiRateLimitRule.MaxQps != nil {
		_ = d.Set("max_qps", apiRateLimitRule.MaxQps)
	}

	if apiRateLimitRule.UsableStatus != nil {
		_ = d.Set("usable_status", apiRateLimitRule.UsableStatus)
	}

	if apiRateLimitRule.RuleId != nil {
		_ = d.Set("rule_id", apiRateLimitRule.RuleId)
	}

	if apiRateLimitRule.RuleName != nil {
		_ = d.Set("rule_name", apiRateLimitRule.RuleName)
	}

	if apiRateLimitRule.RuleContent != nil {
		_ = d.Set("rule_content", apiRateLimitRule.RuleContent)
	}

	if apiRateLimitRule.TsfRuleId != nil {
		_ = d.Set("tsf_rule_id", apiRateLimitRule.TsfRuleId)
	}

	if apiRateLimitRule.Description != nil {
		_ = d.Set("description", apiRateLimitRule.Description)
	}

	if apiRateLimitRule.CreatedTime != nil {
		_ = d.Set("created_time", apiRateLimitRule.CreatedTime)
	}

	if apiRateLimitRule.UpdatedTime != nil {
		_ = d.Set("updated_time", apiRateLimitRule.UpdatedTime)
	}

	return nil
}

func resourceTencentCloudTsfApiRateLimitRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_tsf_api_rate_limit_rule.update")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	request := tsf.NewUpdateApiRateLimitRuleRequest()

	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	// apiId := idSplit[0]
	ruleId := idSplit[1]

	request.RuleId = &ruleId

	immutableArgs := []string{"api_id"}

	for _, v := range immutableArgs {
		if d.HasChange(v) {
			return fmt.Errorf("argument `%s` cannot be changed", v)
		}
	}

	if d.HasChange("max_qps") {
		if v, _ := d.GetOk("max_qps"); v != nil {
			request.MaxQps = helper.IntInt64(v.(int))
		}
	}

	if d.HasChange("usable_status") {
		if v, ok := d.GetOk("usable_status"); ok {
			request.UsableStatus = helper.String(v.(string))
		}
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseTsfClient().UpdateApiRateLimitRule(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s update tsf apiRateLimitRule failed, reason:%+v", logId, err)
		return err
	}

	return resourceTencentCloudTsfApiRateLimitRuleRead(d, meta)
}

func resourceTencentCloudTsfApiRateLimitRuleDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_tsf_api_rate_limit_rule.delete")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := TsfService{client: meta.(*TencentCloudClient).apiV3Conn}
	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	apiId := idSplit[0]
	ruleId := idSplit[1]

	if err := service.DeleteTsfApiRateLimitRuleById(ctx, apiId, ruleId); err != nil {
		return err
	}

	return nil
}
