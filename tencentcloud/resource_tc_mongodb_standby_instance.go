/*
Provide a resource to create a Mongodb standby instance.

Example Usage

```hcl
provider "tencentcloud" {
  region = "ap-guangzhou"
}

provider "tencentcloud" {
  alias  = "shanghai"
  region = "ap-shanghai"
}

resource "tencentcloud_mongodb_instance" "mongodb" {
  instance_name  = "tf-mongodb-test"
  memory         = 4
  volume         = 100
  engine_version = "MONGO_40_WT"
  machine_type   = "HIO10G"
  available_zone = var.availability_zone
  project_id     = 0
  password       = "test1234"

  tags = {
    test = "test"
  }
}

resource "tencentcloud_mongodb_standby_instance" "mongodb" {
  provider               = tencentcloud.shanghai
  instance_name          = "tf-mongodb-standby-test"
  memory                 = 4
  volume                 = 100
  available_zone         = "ap-shanghai-2"
  project_id             = 0
  father_instance_id     = tencentcloud_mongodb_instance.mongodb.id
  father_instance_region = "ap-guangzhou"

  tags = {
    test = "test"
  }
}
```

Import

Mongodb instance can be imported using the id, e.g.

```
$ terraform import tencentcloud_mongodb_standby_instance.mongodb cmgo-41s6jwy4
```
*/
package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	mongodb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

func resourceTencentCloudMongodbStandbyInstance() *schema.Resource {
	mongodbStandbyInstanceInfo := map[string]*schema.Schema{
		"father_instance_region": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			Description: "Indicates the region of main instance.",
		},
		"father_instance_id": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			Description: "Indicates the main instance ID of standby instances.",
		},
		"available_zone": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The available zone of the Mongodb standby instance. NOTE: must not be same with father instance's.",
		},
		// computed from father instance
		"engine_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Version of the standby Mongodb instance and must be same as the version of main instance.",
		},
		"machine_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Type of standby Mongodb instance and must be same as the type of main instance.",
		},
	}
	basic := TencentMongodbBasicInfo()
	conflictList := []string{"engine_version", "machine_type", "password", "available_zone", "mongos_cpu", "mongos_memory", "mongos_node_num"}
	for _, item := range conflictList {
		delete(basic, item)
	}
	for k, v := range basic {
		mongodbStandbyInstanceInfo[k] = v
	}

	return &schema.Resource{
		Create: resourceTencentCloudMongodbStandbyInstanceCreate,
		Read:   resourceTencentCloudMongodbStandbyInstanceRead,
		Update: resourceTencentCloudMongodbStandbyInstanceUpdate,
		Delete: resourceTencentCloudMongodbStandbyInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mongodbStandbyInstanceInfo,
	}
}

func mongodbAllStandbyInstanceReqSet(requestInter interface{}, d *schema.ResourceData, masterInfo map[string]string) error {
	var (
		replicateSetNum       = 1
		nodeNum               = 3
		goodsNum              = 1
		clusterType           = MONGODB_CLUSTER_TYPE_REPLSET
		memoryInterface       = d.Get("memory").(int)
		volumeInterface       = d.Get("volume").(int)
		mongoVersionInterface = masterInfo["engine_version"]
		zoneInterface         = d.Get("available_zone").(string)
		machine               = masterInfo["machine_type"]
		fatherId              = masterInfo["father_instance_id"]
		instanceType          = MONGO_INSTANCE_TYPE_STANDBY
		projectId             = d.Get("project_id").(int)
	)

	getType := reflect.TypeOf(requestInter)
	value := reflect.ValueOf(requestInter).Elem()

	for k, v := range map[string]interface{}{
		"ReplicateSetNum": helper.IntUint64(replicateSetNum),
		"NodeNum":         helper.IntUint64(nodeNum),
		"GoodsNum":        helper.IntUint64(goodsNum),
		"ClusterType":     &clusterType,
		"Memory":          helper.IntUint64(memoryInterface),
		"Volume":          helper.IntUint64(volumeInterface),
		"MongoVersion":    &mongoVersionInterface,
		"Zone":            &zoneInterface,
		"MachineCode":     &machine,
		"Clone":           helper.IntInt64(instanceType),
		"ProjectId":       helper.IntInt64(projectId),
		"Father":          &fatherId,
	} {
		value.FieldByName(k).Set(reflect.ValueOf(v))
	}

	var okVpc, okSubnet bool
	if v, ok := d.GetOk("vpc_id"); ok {
		okVpc = ok
		value.FieldByName("VpcId").Set(reflect.ValueOf(helper.String(v.(string))))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		okSubnet = ok
		value.FieldByName("SubnetId").Set(reflect.ValueOf(helper.String(v.(string))))
	}
	if (okVpc && !okSubnet) || (!okVpc && okSubnet) {
		return fmt.Errorf("you have to set vpc_id and subnet_id both")
	}
	if v, ok := d.GetOk("security_groups"); ok {
		sliceReflect := helper.InterfacesStringsPoint(v.(*schema.Set).List())
		value.FieldByName("SecurityGroup").Set(reflect.ValueOf(sliceReflect))
	}

	if strings.Contains(getType.String(), "CreateDBInstanceRequest") {
		if v, ok := d.GetOk("prepaid_period"); ok {
			value.FieldByName("Period").Set(reflect.ValueOf(helper.IntUint64(v.(int))))
		} else {
			return fmt.Errorf("prepaid_period must be specified for a PREPAID instance")
		}
		value.FieldByName("AutoRenewFlag").Set(reflect.ValueOf(helper.IntUint64(d.Get("auto_renew_flag").(int))))
	}

	return nil
}

func mongodbCreateStandbyInstanceByUse(ctx context.Context, d *schema.ResourceData, meta interface{}, masterInfo map[string]string) error {
	logId := getLogId(ctx)
	request := mongodb.NewCreateDBInstanceHourRequest()

	if err := mongodbAllStandbyInstanceReqSet(request, d, masterInfo); err != nil {
		return err
	}

	var response *mongodb.CreateDBInstanceHourResponse
	var err error
	err = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		response, err = meta.(*TencentCloudClient).apiV3Conn.UseMongodbClient().CreateDBInstanceHour(request)
		if err != nil {
			log.Printf("[CRITAL]%s api[%s] fail, reason:%s", logId, request.GetAction(), err.Error())
			return retryError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(response.Response.InstanceIds) < 1 {
		return fmt.Errorf("mongodb standby instance id is nil")
	}
	d.SetId(*response.Response.InstanceIds[0])

	return nil
}

func mongodbCreateStandbyInstanceByMonth(ctx context.Context, d *schema.ResourceData, meta interface{}, masterInfo map[string]string) error {
	logId := getLogId(ctx)
	request := mongodb.NewCreateDBInstanceRequest()

	if err := mongodbAllStandbyInstanceReqSet(request, d, masterInfo); err != nil {
		return err
	}

	var response *mongodb.CreateDBInstanceResponse
	var err error
	err = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		response, err = meta.(*TencentCloudClient).apiV3Conn.UseMongodbClient().CreateDBInstance(request)
		if err != nil {
			log.Printf("[CRITAL]%s api[%s] fail, reason:%s", logId, request.GetAction(), err.Error())
			return retryError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(response.Response.InstanceIds) < 1 {
		return fmt.Errorf("mongodb standby instance id is nil")
	}
	d.SetId(*response.Response.InstanceIds[0])

	return nil
}

func resourceTencentCloudMongodbStandbyInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_mongodb_standby_instance.create")()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	client := meta.(*TencentCloudClient).apiV3Conn
	client1 := *client
	mongodbService := MongodbService{client: client}
	mongodbService1 := MongodbService{client: &client1}
	tagService := TagService{client: client}
	region := client.Region

	// collect info from master instance
	var masterInfoMap = make(map[string]string, 3)
	if _, ok := d.GetOk("father_instance_id"); !ok {
		return fmt.Errorf("[CRITAL] father instance id must be specified for standby instance")
	}
	if _, ok := d.GetOk("father_instance_region"); !ok {
		return fmt.Errorf("[CRITAL] father instance region must be specified for standby instance")
	}
	fatherRegion := d.Get("father_instance_region").(string)
	mongodbService1.client.Region = fatherRegion
	masterInfoMap["father_instance_id"] = d.Get("father_instance_id").(string)
	masterInfo, has, err := mongodbService1.DescribeInstanceById(ctx, masterInfoMap["father_instance_id"])
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("[CRITAL] father instance can't be found for creating mongodb standby instance")
	}

	masterInfoMap["machine_type"] = *masterInfo.MachineType
	masterInfoMap["engine_version"] = *masterInfo.MongoVersion

	// check available_zone info
	if d.Get("available_zone").(string) == *masterInfo.Zone {
		return fmt.Errorf("[CRITAL] standBy instance zoneId must not same with father instance's")
	}

	// check security group info
	if *masterInfo.MongoVersion == MONGODB_ENGINE_VERSION_4_WT {
		if _, ok := d.GetOk("security_groups"); ok {
			return fmt.Errorf("[CRITAL] for instance which `engine_version` is `MONGO_40_WT`, `security_groups` is not supported")
		}
	}

	chargeType := d.Get("charge_type").(string)

	if chargeType == MONGODB_CHARGE_TYPE_POSTPAID {
		_, ok := d.GetOk("prepaid_period")
		_, ok1 := d.GetOk("auto_renew_flag")
		if ok || ok1 {
			return fmt.Errorf("prepaid_period and auto_renew_flag don't make sense for POSTPAID_BY_HOUR mongodb standby instance, please remove them from your template")
		}
		if err := mongodbCreateStandbyInstanceByUse(ctx, d, meta, masterInfoMap); err != nil {
			return err
		}
	} else {
		if err := mongodbCreateStandbyInstanceByMonth(ctx, d, meta, masterInfoMap); err != nil {
			return err
		}
	}

	instanceId := d.Id()

	_, has, err = mongodbService.DescribeInstanceById(ctx, instanceId)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("[CRITAL]%s creating mongodb standby instance failed, instance doesn't exist", logId)
	}

	// setting instance name
	instanceName := d.Get("instance_name").(string)
	err = mongodbService.ModifyInstanceName(ctx, instanceId, instanceName)
	if err != nil {
		return err
	}

	_, has, err = mongodbService.DescribeInstanceById(ctx, instanceId)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("[CRITAL]%s creating mongodb instance failed, instance doesn't exist", logId)
	}

	if tags := helper.GetTags(d, "tags"); len(tags) > 0 {
		resourceName := BuildTagResourceName("mongodb", "instance", region, instanceId)
		if err := tagService.ModifyTags(ctx, resourceName, tags, nil); err != nil {
			return err
		}
	}

	return resourceTencentCloudMongodbInstanceRead(d, meta)
}

func resourceTencentCloudMongodbStandbyInstanceRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_mongodb_standby_instance.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	instanceId := d.Id()

	mongodbService := MongodbService{client: meta.(*TencentCloudClient).apiV3Conn}
	instance, has, err := mongodbService.DescribeInstanceById(ctx, instanceId)
	if err != nil {
		return err
	}
	if !has {
		d.SetId("")
		return nil
	}

	if nilFields := CheckNil(instance, map[string]string{
		"InstanceName": "instance name",
		"ProjectId":    "project id",
		"Zone":         "available zone",
		"VpcId":        "vpc id",
		"SubnetId":     "subnet id",
		"Status":       "status",
		"Vip":          "vip",
		"Vport":        "vport",
		"CreateTime":   "create time",
		"MongoVersion": "engine version",
		"Memory":       "memory",
		"Volume":       "volume",
		"MachineType":  "machine type",
	}); len(nilFields) > 0 {
		return fmt.Errorf("mongodb %v are nil", nilFields)
	}

	_ = d.Set("instance_name", instance.InstanceName)
	_ = d.Set("memory", *instance.Memory/1024)
	_ = d.Set("volume", *instance.Volume/1024)
	_ = d.Set("engine_version", instance.MongoVersion)
	_ = d.Set("charge_type", MONGODB_CHARGE_TYPE[*instance.PayMode])
	if MONGODB_CHARGE_TYPE[*instance.PayMode] == MONGODB_CHARGE_TYPE_PREPAID {
		_ = d.Set("auto_renew_flag", *instance.AutoRenewFlag)
	}

	_ = d.Set("machine_type", *instance.MachineType)
	_ = d.Set("available_zone", instance.Zone)
	_ = d.Set("vpc_id", instance.VpcId)
	_ = d.Set("subnet_id", instance.SubnetId)
	_ = d.Set("project_id", instance.ProjectId)
	_ = d.Set("status", instance.Status)
	_ = d.Set("vip", instance.Vip)
	_ = d.Set("vport", instance.Vport)
	_ = d.Set("create_time", instance.CreateTime)

	// info of master info
	if instance.RelatedInstance != nil {
		_ = d.Set("father_instance_id", instance.RelatedInstance.InstanceId)
		_ = d.Set("father_instance_region", instance.RelatedInstance.Region)
	}

	tags := make(map[string]string, len(instance.Tags))
	for _, tag := range instance.Tags {
		if tag.TagKey == nil {
			return errors.New("mongodb tag key is nil")
		}
		if tag.TagValue == nil {
			return errors.New("mongodb tag value is nil")
		}
		if *tag.TagKey == "project" {
			continue
		}

		tags[*tag.TagKey] = *tag.TagValue
	}
	_ = d.Set("tags", tags)

	return nil
}

func resourceTencentCloudMongodbStandbyInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_mongodb_standby_instance.update")()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	instanceId := d.Id()

	client := meta.(*TencentCloudClient).apiV3Conn
	mongodbService := MongodbService{client: client}
	tagService := TagService{client: client}
	region := client.Region

	d.Partial(true)

	if d.HasChange("memory") || d.HasChange("volume") {
		memory := d.Get("memory").(int)
		volume := d.Get("volume").(int)
		err := mongodbService.UpgradeInstance(ctx, instanceId, memory, volume)
		if err != nil {
			return err
		}

		// it will take time to wait for memory and volume change even describe request succeeded even the status returned in describe response is running
		errUpdate := resource.Retry(20*readRetryTimeout, func() *resource.RetryError {
			infos, has, e := mongodbService.DescribeInstanceById(ctx, instanceId)
			if e != nil {
				return resource.NonRetryableError(e)
			}
			if !has {
				return resource.NonRetryableError(fmt.Errorf("[CRITAL]%s updating mongodb instance failed, instance doesn't exist", logId))
			}

			memoryDes := *infos.Memory / 1024 / (*infos.ReplicationSetNum)
			volumeDes := *infos.Volume / 1024 / (*infos.ReplicationSetNum)
			if memory != int(memoryDes) || volume != int(volumeDes) {
				return resource.RetryableError(fmt.Errorf("[CRITAL] updating mongodb instance, current memory and volume values: %d, %d, waiting for them becoming new value: %d, %d", memoryDes, volumeDes, d.Get("memory").(int), d.Get("volume").(int)))
			}
			return nil
		})
		if errUpdate != nil {
			return errUpdate
		}

		d.SetPartial("memory")
		d.SetPartial("volume")
	}

	if d.HasChange("instance_name") {
		instanceName := d.Get("instance_name").(string)
		err := mongodbService.ModifyInstanceName(ctx, instanceId, instanceName)
		if err != nil {
			return err
		}
		d.SetPartial("instance_name")
	}

	if d.HasChange("project_id") {
		projectId := d.Get("project_id").(int)
		err := mongodbService.ModifyProjectId(ctx, instanceId, projectId)
		if err != nil {
			return err
		}
		d.SetPartial("project_id")
	}

	if d.HasChange("tags") {
		oldTags, newTags := d.GetChange("tags")
		replaceTags, deleteTags := diffTags(oldTags.(map[string]interface{}), newTags.(map[string]interface{}))

		resourceName := BuildTagResourceName("mongodb", "instance", region, instanceId)
		if err := tagService.ModifyTags(ctx, resourceName, replaceTags, deleteTags); err != nil {
			return err
		}

		d.SetPartial("tags")
	}

	if d.HasChange("prepaid_period") {
		return fmt.Errorf("setting of the field[prepaid_period] does not make sense after the initialization")
	}

	if d.HasChange("auto_renew_flag") {
		autoRenewFlag := d.Get("auto_renew_flag").(int)
		period := d.Get("prepaid_period").(int)
		err := mongodbService.ModifyAutoRenewFlag(ctx, instanceId, period, autoRenewFlag)
		if err != nil {
			return err
		}
		d.SetPartial("auto_renew_flag")
	}

	d.Partial(false)

	return resourceTencentCloudMongodbStandbyInstanceRead(d, meta)
}

func resourceTencentCloudMongodbStandbyInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_mongodb_standby_instance.delete")()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	instanceId := d.Id()
	mongodbService := MongodbService{
		client: meta.(*TencentCloudClient).apiV3Conn,
	}

	instanceDetail, has, err := mongodbService.DescribeInstanceById(ctx, instanceId)
	if err != nil {
		return err
	}
	if !has {
		d.SetId("")
		return nil
	}
	if MONGODB_CHARGE_TYPE[*instanceDetail.PayMode] == MONGODB_CHARGE_TYPE_PREPAID {
		return fmt.Errorf("PREPAID instances are not allowed to be deleted now, please isolate them on console")
	}

	err = mongodbService.IsolateInstance(ctx, instanceId)
	if err != nil {
		return err
	}
	err = mongodbService.OfflineIsolatedDBInstance(ctx, instanceId, true)
	if err != nil {
		log.Printf("[CRITAL]%s mongodb %s fail, reason:%s", logId, "OfflineIsolatedDBInstance", err.Error())
		return err
	}
	//describe and check not exist
	_, has, errRet := mongodbService.DescribeInstanceById(ctx, instanceId)
	if errRet != nil {
		return errRet
	}
	if !has {
		return nil
	}
	return fmt.Errorf("[CRITAL]%s mongodb %s fail", logId, "OfflineIsolatedDBInstance")
}
