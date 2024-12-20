package gaussdb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

const (
	haModeDistributed = "enterprise"
	haModeCentralized = "centralization_standard"
)

// @API GaussDB GET /v3/{project_id}/instances
// @API GaussDB POST /v3/{project_id}/instances
// @API GaussDB GET /v3/{project_id}/jobs
// @API GaussDB PUT /v3/{project_id}/instances/{instance_id}/name
// @API GaussDB POST /v3/{project_id}/instances/{instance_id}/password
// @API GaussDB POST /v3/{project_id}/instances/{instance_id}/action
// @API GaussDB PUT /v3/{project_id}/instances/{instance_id}/backups/policy
// @API GaussDB DELETE /v3/{project_id}/instances/{instance_id}
// @API BSS GET /v2/orders/customer-orders/details/{order_id}
// @API BSS POST /v2/orders/suscriptions/resources/query
// @API BSS POST /v2/orders/subscriptions/resources/autorenew/{instance_id}
// @API BSS DELETE /v2/orders/subscriptions/resources/autorenew/{instance_id}
// @API BSS POST /v2/orders/subscriptions/resources/unsubscribe
func ResourceOpenGaussInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenGaussInstanceCreate,
		ReadContext:   resourceOpenGaussInstanceRead,
		UpdateContext: resourceOpenGaussInstanceUpdate,
		DeleteContext: resourceOpenGaussInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, v interface{}) error {
			if d.HasChange("coordinator_num") {
				return d.SetNewComputed("private_ips")
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ha": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: utils.SuppressCaseDiffs,
						},
						"replication_mode": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"consistency": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Computed:         true,
							DiffSuppressFunc: utils.SuppressCaseDiffs,
						},
						"instance_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"sharding_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"coordinator_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"replica_num": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  3,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"configuration_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"port": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enterprise_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disk_encryption_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_force_switch": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_single_float_ip": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "UTC+08:00",
			},
			"datastore": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"engine": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"backup_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keep_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"force_import": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"charging_mode": common.SchemaChargingMode(nil),
			"period_unit":   common.SchemaPeriodUnit(nil),
			"period":        common.SchemaPeriod(nil),
			"auto_renew":    common.SchemaAutoRenewUpdatable(nil),

			// Attributes
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"db_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"switch_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceOpenGaussInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		httpUrl = "v3/{project_id}/instances"
		product = "opengauss"
	)
	client, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating GaussDB client: %s", err)
	}

	// If force_import set, try to import it instead of creating
	if common.HasFilledOpt(d, "force_import") {
		log.Printf("[DEBUG] the Gaussdb OpenGauss instance force_import is set, try to import it instead of creating")
		instanceName := d.Get("name").(string)
		res, err := getGaussDBOpenGaussInstancesByName(client, instanceName)
		if err != nil {
			return diag.FromErr(err)
		}
		instanceId := utils.PathSearch("instances[0].id", res, nil)
		if instanceId == nil {
			return diag.Errorf("unable to retrieve instances by name: %s", instanceName)
		}
		log.Printf("[DEBUG] found existing OpenGauss instance %v with name %s", instanceId, instanceName)
		d.SetId(instanceId.(string))
		return resourceOpenGaussInstanceRead(ctx, d, meta)
	}

	createPath := client.Endpoint + httpUrl
	createPath = strings.ReplaceAll(createPath, "{project_id}", client.ProjectID)

	createOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	createOpt.JSONBody = utils.RemoveNil(buildCreateGaussDBOpenGaussBodyParams(d, region))
	createResp, err := client.Request("POST", createPath, &createOpt)
	if err != nil {
		return diag.Errorf("error creating GaussDB OpenGauss instance: %s", err)
	}
	createRespBody, err := utils.FlattenResponse(createResp)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceId := utils.PathSearch("instance.id", createRespBody, nil)
	if instanceId == nil {
		return diag.Errorf("error creating GaussDB OpenGauss instance: id is not found in API response")
	}
	d.SetId(instanceId.(string))

	if v, ok := d.GetOk("charging_mode"); ok && v.(string) == "prePaid" {
		orderId := utils.PathSearch("order_id", createRespBody, nil)
		if orderId == nil {
			return diag.Errorf("error creating GaussDB OpenGauss instance: order_id is not found in API response")
		}
		bssClient, err := cfg.BssV2Client(region)
		if err != nil {
			return diag.Errorf("error creating BSS v2 client: %s", err)
		}
		// wait for order success
		err = common.WaitOrderComplete(ctx, bssClient, orderId.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = common.WaitOrderResourceComplete(ctx, bssClient, orderId.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.Errorf("error waiting for GaussDB OpenGauss order %s complete: %s", orderId, err)
		}
	} else {
		jobId := utils.PathSearch("job_id", createRespBody, nil)
		if jobId == nil {
			return diag.Errorf("error creating GaussDB OpenGauss instance: job_id is not found in API response")
		}
		err = checkGaussDBOpenGaussJobFinish(ctx, client, jobId.(string), 300, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"BUILD", "BACKING UP"},
		Target:       []string{"ACTIVE"},
		Refresh:      instanceStateRefreshFunc(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for instance (%s) to become ready: %s", d.Id(), err)
	}

	// This is a workaround to avoid db connection issue
	time.Sleep(360 * time.Second) // lintignore:R018

	return resourceOpenGaussInstanceRead(ctx, d, meta)
}

func getGaussDBOpenGaussInstancesById(client *golangsdk.ServiceClient, id string) (interface{}, error) {
	queryParam := fmt.Sprintf("?id=%s", id)
	return getGaussDBOpenGaussInstances(client, queryParam)
}

func getGaussDBOpenGaussInstancesByName(client *golangsdk.ServiceClient, name string) (interface{}, error) {
	queryParam := fmt.Sprintf("?name=%s", name)
	return getGaussDBOpenGaussInstances(client, queryParam)
}

func getGaussDBOpenGaussInstances(client *golangsdk.ServiceClient, queryParam string) (interface{}, error) {
	var (
		httpUrl = "v3/{project_id}/instances"
	)

	getPath := client.Endpoint + httpUrl
	getPath = strings.ReplaceAll(getPath, "{project_id}", client.ProjectID)
	getPath += queryParam

	getOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	getResp, err := client.Request("GET", getPath, &getOpt)
	if err != nil {
		return nil, err
	}

	getRespBody, err := utils.FlattenResponse(getResp)
	if err != nil {
		return nil, err
	}
	return getRespBody, nil
}

func buildCreateGaussDBOpenGaussBodyParams(d *schema.ResourceData, region string) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"name":                  d.Get("name"),
		"datastore":             buildCreateGaussDBOpenGaussDatastoreBodyParams(d),
		"ha":                    buildCreateGaussDBOpenGaussHaBodyParams(d),
		"configuration_id":      utils.ValueIgnoreEmpty(d.Get("configuration_id")),
		"port":                  utils.ValueIgnoreEmpty(d.Get("port")),
		"password":              d.Get("password"),
		"backup_strategy":       buildCreateGaussDBOpenGaussBackupStrategyBodyParams(d),
		"enterprise_project_id": utils.ValueIgnoreEmpty(d.Get("enterprise_project_id")),
		"disk_encryption_id":    utils.ValueIgnoreEmpty(d.Get("disk_encryption_id")),
		"flavor_ref":            d.Get("flavor"),
		"volume":                buildCreateGaussDBOpenGaussVolumeBodyParams(d),
		"region":                region,
		"availability_zone":     utils.ValueIgnoreEmpty(d.Get("availability_zone")),
		"vpc_id":                d.Get("vpc_id"),
		"subnet_id":             d.Get("subnet_id"),
		"security_group_id":     utils.ValueIgnoreEmpty(d.Get("security_group_id")),
		"charge_info":           buildCreateGaussDBOpenGaussChargeInfoBodyParams(d),
		"time_zone":             utils.ValueIgnoreEmpty(d.Get("time_zone")),
		"sharding_num":          utils.ValueIgnoreEmpty(d.Get("sharding_num")),
		"coordinator_num":       utils.ValueIgnoreEmpty(d.Get("coordinator_num")),
		"replica_num":           utils.ValueIgnoreEmpty(d.Get("replica_num")),
	}
	if v := d.Get("enable_force_switch").(bool); v {
		bodyParams["enable_force_switch"] = v
	}
	if v := d.Get("enable_single_float_ip").(bool); v {
		bodyParams["enable_single_float_ip"] = v
	}
	return bodyParams
}

func buildCreateGaussDBOpenGaussDatastoreBodyParams(d *schema.ResourceData) map[string]interface{} {
	rawParams := d.Get("datastore").([]interface{})
	if len(rawParams) > 0 {
		rawParam := rawParams[0].(map[string]interface{})
		return map[string]interface{}{
			"type":    rawParam["engine"],
			"version": rawParam["version"],
		}
	}
	return map[string]interface{}{
		"type": "GaussDB(for openGauss)",
	}
}

func buildCreateGaussDBOpenGaussHaBodyParams(d *schema.ResourceData) map[string]interface{} {
	rawParams := d.Get("ha").([]interface{})
	rawParam := rawParams[0].(map[string]interface{})
	bodyParams := map[string]interface{}{
		"mode":             rawParam["mode"],
		"replication_mode": rawParam["replication_mode"],
		"consistency":      rawParam["consistency"],
		"instance_mode":    utils.ValueIgnoreEmpty(rawParam["instance_mode"]),
	}
	return bodyParams
}

func buildCreateGaussDBOpenGaussBackupStrategyBodyParams(d *schema.ResourceData) map[string]interface{} {
	rawParams := d.Get("backup_strategy").([]interface{})
	if len(rawParams) == 0 {
		return nil
	}
	rawParam := rawParams[0].(map[string]interface{})
	bodyParams := map[string]interface{}{
		"start_time": rawParam["start_time"],
		"keep_days":  utils.ValueIgnoreEmpty(rawParam["keep_days"]),
	}
	return bodyParams
}

func buildCreateGaussDBOpenGaussVolumeBodyParams(d *schema.ResourceData) map[string]interface{} {
	rawParams := d.Get("volume").([]interface{})
	mode := d.Get("ha").([]interface{})[0].(map[string]interface{})["mode"].(string)
	dnNum := 1
	if mode == haModeDistributed {
		dnNum = d.Get("sharding_num").(int)
	}
	if mode == haModeCentralized {
		dnNum = d.Get("replica_num").(int) + 1
	}
	raw := rawParams[0].(map[string]interface{})
	dnSize := raw["size"].(int)
	bodyParams := map[string]interface{}{
		"type": raw["type"],
		"size": dnSize * dnNum,
	}
	return bodyParams
}

func buildCreateGaussDBOpenGaussChargeInfoBodyParams(d *schema.ResourceData) map[string]interface{} {
	if v, ok := d.GetOk("charging_mode"); ok && v == "prePaid" {
		bodyParams := map[string]interface{}{
			"charge_mode":   utils.ValueIgnoreEmpty(d.Get("charging_mode")),
			"period_type":   utils.ValueIgnoreEmpty(d.Get("period_unit")),
			"period_num":    utils.ValueIgnoreEmpty(d.Get("period")),
			"is_auto_renew": d.Get("auto_renew"),
			"is_auto_pay":   "true",
		}
		return bodyParams
	}
	return nil
}

func resourceOpenGaussInstanceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		product = "opengauss"
	)
	client, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating GaussDB client: %s", err)
	}

	res, err := getGaussDBOpenGaussInstancesById(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	instance := utils.PathSearch("instances[0]", res, nil)
	if instance == nil {
		return common.CheckDeletedDiag(d, golangsdk.ErrDefault404{}, "")
	}

	var dnNum = 1
	mErr := multierror.Append(
		d.Set("region", region),
		d.Set("name", utils.PathSearch("name", instance, nil)),
		d.Set("status", utils.PathSearch("status", instance, nil)),
		d.Set("type", utils.PathSearch("type", instance, nil)),
		d.Set("vpc_id", utils.PathSearch("vpc_id", instance, nil)),
		d.Set("subnet_id", utils.PathSearch("subnet_id", instance, nil)),
		d.Set("security_group_id", utils.PathSearch("security_group_id", instance, nil)),
		d.Set("db_user_name", utils.PathSearch("db_user_name", instance, nil)),
		d.Set("time_zone", utils.PathSearch("time_zone", instance, nil)),
		d.Set("flavor", utils.PathSearch("flavor_ref", instance, nil)),
		d.Set("port", strconv.Itoa(int(utils.PathSearch("port", instance, float64(0)).(float64)))),
		d.Set("switch_strategy", utils.PathSearch("switch_strategy", instance, nil)),
		d.Set("maintenance_window", utils.PathSearch("maintenance_window", instance, nil)),
		d.Set("public_ips", utils.PathSearch("public_ips", instance, nil)),
		d.Set("charging_mode", utils.PathSearch("charge_info.charge_mode", instance, nil)),
		d.Set("ha", flattenGaussDBOpenGaussResponseBodyHa(instance)),
		d.Set("datastore", flattenGaussDBOpenGaussResponseBodyDatastore(instance)),
		d.Set("backup_strategy", flattenGaussDBOpenGaussResponseBodyBackupStrategy(instance)),
		setOpenGaussNodesAndRelatedNumbers(d, instance, &dnNum),
		d.Set("volume", flattenGaussDBOpenGaussResponseBodyVolume(instance, dnNum)),
		setOpenGaussPrivateIpsAndEndpoints(d, instance),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenGaussDBOpenGaussResponseBodyHa(instance interface{}) []interface{} {
	rst := []interface{}{
		map[string]interface{}{
			"mode":             utils.PathSearch("type", instance, nil),
			"replication_mode": utils.PathSearch("ha.replication_mode", instance, nil),
			"consistency":      utils.PathSearch("ha.consistency", instance, nil),
			"instance_mode":    utils.PathSearch("instance_mode", instance, nil),
		},
	}
	return rst
}

func flattenGaussDBOpenGaussResponseBodyDatastore(instance interface{}) []interface{} {
	rst := []interface{}{
		map[string]interface{}{
			"engine":  utils.PathSearch("datastore.type", instance, nil),
			"version": utils.PathSearch("datastore.version", instance, nil),
		},
	}
	return rst
}

func flattenGaussDBOpenGaussResponseBodyBackupStrategy(instance interface{}) []interface{} {
	rst := []interface{}{
		map[string]interface{}{
			"start_time": utils.PathSearch("backup_strategy.start_time", instance, nil),
			"keep_days":  utils.PathSearch("backup_strategy.keep_days", instance, nil),
		},
	}
	return rst
}

func flattenGaussDBOpenGaussResponseBodyVolume(instance interface{}, dnNum int) []interface{} {
	rst := []interface{}{
		map[string]interface{}{
			"type": utils.PathSearch("volume.type", instance, nil),
			"size": int(utils.PathSearch("volume.size", instance, float64(0)).(float64)) / dnNum,
		},
	}
	return rst
}

func setOpenGaussNodesAndRelatedNumbers(d *schema.ResourceData, instance interface{}, dnNum *int) *multierror.Error {
	shardingNum := 0
	coordinatorNum := 0

	nodesArray := utils.PathSearch("nodes", instance, make([]interface{}, 0)).([]interface{})
	nodesList := make([]map[string]interface{}, 0, len(nodesArray))
	for _, v := range nodesArray {
		name := utils.PathSearch("name", v, "").(string)
		node := map[string]interface{}{
			"id":                utils.PathSearch("id", v, nil),
			"name":              name,
			"status":            utils.PathSearch("status", v, nil),
			"role":              utils.PathSearch("role", v, nil),
			"availability_zone": utils.PathSearch("availability_zone", v, nil),
		}
		nodesList = append(nodesList, node)

		if strings.Contains(name, "_gaussdbv5cn") {
			coordinatorNum++
		} else if strings.Contains(name, "_gaussdbv5dn") {
			shardingNum++
		}
	}

	if shardingNum > 0 && coordinatorNum > 0 {
		*dnNum = shardingNum / d.Get("replica_num").(int)
		mErr := multierror.Append(
			d.Set("nodes", nodesList),
			d.Set("sharding_num", dnNum),
			d.Set("coordinator_num", coordinatorNum),
		)
		return mErr
	}

	// If the HA mode is centralized, the HA structure of API response is nil.
	replicaNum := utils.PathSearch("replica_num", instance, float64(0)).(float64)
	*dnNum = int(replicaNum) + 1
	mErr := multierror.Append(
		d.Set("nodes", nodesList),
		d.Set("replica_num", replicaNum),
	)
	return mErr
}

func setOpenGaussPrivateIpsAndEndpoints(d *schema.ResourceData, instance interface{}) *multierror.Error {
	privateIps := utils.PathSearch("private_ips", instance, make([]interface{}, 0)).([]interface{})
	if len(privateIps) == 0 {
		return nil
	}

	port := utils.PathSearch("port", instance, float64(0)).(float64)
	ipList := strings.Split(privateIps[0].(string), "/")
	endpoints := make([]string, 0, len(ipList))
	for i := 0; i < len(ipList); i++ {
		ipList[i] = strings.Trim(ipList[i], " ")
		endpoint := fmt.Sprintf("%s:%v", ipList[i], port)
		endpoints = append(endpoints, endpoint)
	}

	mErr := multierror.Append(
		d.Set("private_ips", ipList),
		d.Set("endpoints", endpoints),
	)

	return mErr
}

func resourceOpenGaussInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		product = "opengauss"
	)
	client, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating GaussDB client: %s", err)
	}
	bssClient, err := cfg.BssV2Client(region)
	if err != nil {
		return diag.Errorf("error creating BSS v2 client: %s", err)
	}

	if d.HasChange("name") {
		if err = updateInstanceName(ctx, d, client); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		if err = updateInstancePassword(d, client); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("sharding_num") {
		if err = expandInstanceShardingNumber(ctx, d, client, bssClient); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("coordinator_num") {
		if err = expandInstanceCoordinatorNumber(ctx, d, client, bssClient); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("volume") {
		if err = updateInstanceVolumeSize(ctx, d, client, bssClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("backup_strategy") {
		if err = updateInstanceBackupStrategy(d, client); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("auto_renew") {
		if err = common.UpdateAutoRenew(bssClient, d.Get("auto_renew").(string), d.Id()); err != nil {
			return diag.Errorf("error updating the auto-renew of the instance (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enterprise_project_id") {
		migrateOpts := config.MigrateResourceOpts{
			ResourceId:   d.Id(),
			ResourceType: "gaussdb",
			RegionId:     region,
			ProjectId:    cfg.GetProjectID(region),
		}
		if err = cfg.MigrateEnterpriseProject(ctx, d, migrateOpts); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceOpenGaussInstanceRead(ctx, d, meta)
}

func updateInstanceName(ctx context.Context, d *schema.ResourceData, client *golangsdk.ServiceClient) error {
	var (
		httpUrl = "v3/{project_id}/instances/{instance_id}/name"
	)

	updatePath := client.Endpoint + httpUrl
	updatePath = strings.ReplaceAll(updatePath, "{project_id}", client.ProjectID)
	updatePath = strings.ReplaceAll(updatePath, "{instance_id}", d.Id())

	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildUpdateInstanceNameBodyParams(d))

	updateResp, err := client.Request("PUT", updatePath, &updateOpt)
	if err != nil {
		return fmt.Errorf("error updating GaussDB OpenGauss instance (%s) name: %s", d.Id(), err)
	}

	updateRespBody, err := utils.FlattenResponse(updateResp)
	if err != nil {
		return err
	}
	jobId := utils.PathSearch("job_id", updateRespBody, nil)
	if jobId == nil {
		return fmt.Errorf("error updating GaussDB OpenGauss instance name: job_id is not found in API response")
	}

	err = checkGaussDBOpenGaussJobFinish(ctx, client, jobId.(string), 2, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}

	return nil
}

func buildUpdateInstanceNameBodyParams(d *schema.ResourceData) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"name": d.Get("name"),
	}
	return bodyParams
}

func updateInstancePassword(d *schema.ResourceData, client *golangsdk.ServiceClient) error {
	var (
		httpUrl = "v3/{project_id}/instances/{instance_id}/password"
	)

	updatePath := client.Endpoint + httpUrl
	updatePath = strings.ReplaceAll(updatePath, "{project_id}", client.ProjectID)
	updatePath = strings.ReplaceAll(updatePath, "{instance_id}", d.Id())

	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildUpdateInstancePasswordBodyParams(d))

	_, err := client.Request("POST", updatePath, &updateOpt)
	if err != nil {
		return fmt.Errorf("error updating GaussDB OpenGauss instance (%s) password: %s", d.Id(), err)
	}

	return nil
}

func buildUpdateInstancePasswordBodyParams(d *schema.ResourceData) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"password": d.Get("password"),
	}
	return bodyParams
}

func expandInstanceShardingNumber(ctx context.Context, d *schema.ResourceData, client, bssClient *golangsdk.ServiceClient) error {
	oRaw, nRaw := d.GetChange("sharding_num")
	if nRaw.(int) < oRaw.(int) {
		return fmt.Errorf("error expanding shard for instance: new num must be larger than the old one")
	}
	expandSize := nRaw.(int) - oRaw.(int)
	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildExpandInstanceShardingNumberBodyParams(expandSize))
	return updateInstanceVolumeAndRelatedHaNumbers(ctx, client, bssClient, d, updateOpt)
}

func buildExpandInstanceShardingNumberBodyParams(expandSize int) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"expand_cluster": map[string]interface{}{
			"shard": map[string]interface{}{
				"count": expandSize,
			},
		},
		"is_auto_pay": "true",
	}
	return bodyParams
}

func expandInstanceCoordinatorNumber(ctx context.Context, d *schema.ResourceData, client, bssClient *golangsdk.ServiceClient) error {
	oRaw, nRaw := d.GetChange("coordinator_num")
	if nRaw.(int) < oRaw.(int) {
		return fmt.Errorf("error expanding coordinator for instance: new number must be larger than the old one")
	}
	expandSize := nRaw.(int) - oRaw.(int)

	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildExpandInstanceCoordinatorNumberBodyParams(d, expandSize))
	return updateInstanceVolumeAndRelatedHaNumbers(ctx, client, bssClient, d, updateOpt)
}

func buildExpandInstanceCoordinatorNumberBodyParams(d *schema.ResourceData, expandSize int) map[string]interface{} {
	coordinators := make([]interface{}, 0)
	azList := strings.Split(d.Get("availability_zone").(string), ",")
	for i := 0; i < expandSize; i++ {
		coordinators = append(coordinators, map[string]interface{}{
			"az_code": azList[0],
		})
	}
	bodyParams := map[string]interface{}{
		"expand_cluster": map[string]interface{}{
			"coordinators": coordinators,
		},
		"is_auto_pay": "true",
	}
	return bodyParams
}

func updateInstanceVolumeSize(ctx context.Context, d *schema.ResourceData, client, bssClient *golangsdk.ServiceClient) error {
	volumeRaw := d.Get("volume").([]interface{})
	dnSize := volumeRaw[0].(map[string]interface{})["size"].(int)
	dnNum := 1
	if d.Get("ha.0.mode").(string) == haModeDistributed {
		dnNum = d.Get("sharding_num").(int)
	}
	if d.Get("ha.0.mode").(string) == haModeCentralized {
		dnNum = d.Get("replica_num").(int) + 1
	}

	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildUpdateInstanceVolumeSizeBodyParams(dnSize, dnNum))
	return updateInstanceVolumeAndRelatedHaNumbers(ctx, client, bssClient, d, updateOpt)
}

func buildUpdateInstanceVolumeSizeBodyParams(dnSize, dnNum int) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"enlarge_volume": map[string]interface{}{
			"size": dnSize * dnNum,
		},
		"is_auto_pay": "true",
	}
	return bodyParams
}

func updateInstanceVolumeAndRelatedHaNumbers(ctx context.Context, client, bssClient *golangsdk.ServiceClient,
	d *schema.ResourceData, opts golangsdk.RequestOpts) error {
	var (
		httpUrl = "v3/{project_id}/instances/{instance_id}/action"
	)

	updatePath := client.Endpoint + httpUrl
	updatePath = strings.ReplaceAll(updatePath, "{project_id}", client.ProjectID)
	updatePath = strings.ReplaceAll(updatePath, "{instance_id}", d.Id())

	updateResp, err := client.Request("POST", updatePath, &opts)
	if err != nil {
		return fmt.Errorf("error updating GaussDB OpenGauss instance (%s): %s", d.Id(), err)
	}

	updateRespBody, err := utils.FlattenResponse(updateResp)
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("charging_mode"); ok && v.(string) == "prePaid" {
		orderId := utils.PathSearch("orderId", updateRespBody, nil)
		if orderId == nil {
			return fmt.Errorf("error updating GaussDB OpenGauss instance: order_id is not found in API response")
		}
		// wait for order success
		err = common.WaitOrderComplete(ctx, bssClient, orderId.(string), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}
	} else {
		jobId := utils.PathSearch("job_id", updateRespBody, nil)
		if jobId == nil {
			return fmt.Errorf("error updating GaussDB OpenGauss instance: job_id is not found in API response")
		}
		err = checkGaussDBOpenGaussJobFinish(ctx, client, jobId.(string), 180, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"MODIFYING", "BACKING UP"},
		Target:       []string{"ACTIVE"},
		Refresh:      instanceStateRefreshFunc(client, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for instance (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}

func updateInstanceBackupStrategy(d *schema.ResourceData, client *golangsdk.ServiceClient) error {
	var (
		httpUrl = "v3/{project_id}/instances/{instance_id}/backups/policy"
	)

	updatePath := client.Endpoint + httpUrl
	updatePath = strings.ReplaceAll(updatePath, "{project_id}", client.ProjectID)
	updatePath = strings.ReplaceAll(updatePath, "{instance_id}", d.Id())

	updateOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	updateOpt.JSONBody = utils.RemoveNil(buildUpdateInstanceBackupStrategyBodyParams(d))

	_, err := client.Request("PUT", updatePath, &updateOpt)
	if err != nil {
		return fmt.Errorf("error updating GaussDB OpenGauss instance (%s) backup strategy: %s", d.Id(), err)
	}

	return nil
}

func buildUpdateInstanceBackupStrategyBodyParams(d *schema.ResourceData) map[string]interface{} {
	rawParams := d.Get("backup_strategy").([]interface{})
	rawParam := rawParams[0].(map[string]interface{})
	params := map[string]interface{}{
		"start_time":          rawParam["start_time"],
		"keep_days":           rawParam["keep_days"],
		"period":              "1,2,3,4,5,6,7",
		"differential_period": "30",
	}
	bodyParams := map[string]interface{}{
		"backup_policy": params,
	}
	return bodyParams
}

func resourceOpenGaussInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		httpUrl = "v3/{project_id}/instances/{instance_id}"
		product = "opengauss"
	)
	client, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating GaussDB client: %s", err)
	}
	if v, ok := d.GetOk("charging_mode"); ok && v.(string) == "prePaid" {
		if err = common.UnsubscribePrePaidResource(d, cfg, []string{d.Id()}); err != nil {
			return diag.Errorf("error unsubscribe OpenGauss instance: %s", err)
		}
	} else {
		deletePath := client.Endpoint + httpUrl
		deletePath = strings.ReplaceAll(deletePath, "{project_id}", client.ProjectID)
		deletePath = strings.ReplaceAll(deletePath, "{instance_id}", d.Id())

		deleteOpt := golangsdk.RequestOpts{
			KeepResponseBody: true,
		}

		deleteResp, err := client.Request("DELETE", deletePath, &deleteOpt)
		if err != nil {
			return common.CheckDeletedDiag(d, err, "error deleting GaussDB OpenGauss instance")
		}

		deleteRespBody, err := utils.FlattenResponse(deleteResp)
		if err != nil {
			return diag.FromErr(err)
		}

		jobId := utils.PathSearch("job_id", deleteRespBody, nil)
		if jobId == nil {
			return diag.Errorf("error deleting GaussDB OpenGauss instance: job_id is not found in API response")
		}

		err = checkGaussDBOpenGaussJobFinish(ctx, client, jobId.(string), 5, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func checkGaussDBOpenGaussJobFinish(ctx context.Context, client *golangsdk.ServiceClient, jobID string, delay int,
	timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Running"},
		Target:       []string{"Completed"},
		Refresh:      gaussDBOpenGaussStatusRefreshFunc(client, jobID),
		Timeout:      timeout,
		Delay:        time.Duration(delay) * time.Second,
		PollInterval: 10 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for GaussDB Opengauss job (%s) to be completed: %s ", jobID, err)
	}
	return nil
}

func instanceStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := getGaussDBOpenGaussInstancesById(client, instanceID)
		if err != nil {
			return nil, "", err
		}
		instance := utils.PathSearch("instances[0]", res, nil)
		if instance == nil {
			return struct{}{}, "DELETED", nil
		}
		status := utils.PathSearch("status", instance, "").(string)
		return instance, status, nil
	}
}

func gaussDBOpenGaussStatusRefreshFunc(client *golangsdk.ServiceClient, jobId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var (
			getJobStatusHttpUrl = "v3/{project_id}/jobs?id={job_id}"
		)

		getJobStatusPath := client.Endpoint + getJobStatusHttpUrl
		getJobStatusPath = strings.ReplaceAll(getJobStatusPath, "{project_id}", client.ProjectID)
		getJobStatusPath = strings.ReplaceAll(getJobStatusPath, "{job_id}", jobId)

		getJobStatusOpt := golangsdk.RequestOpts{
			KeepResponseBody: true,
			MoreHeaders:      map[string]string{"Content-Type": "application/json"},
		}
		getJobStatusResp, err := client.Request("GET", getJobStatusPath, &getJobStatusOpt)
		if err != nil {
			return nil, "Failed", err
		}

		getJobStatusRespBody, err := utils.FlattenResponse(getJobStatusResp)
		if err != nil {
			return nil, "", err
		}

		status := utils.PathSearch("job.status", getJobStatusRespBody, "")
		return getJobStatusRespBody, status.(string), nil
	}
}
