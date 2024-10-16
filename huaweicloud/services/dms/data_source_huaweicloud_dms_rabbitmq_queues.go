// Generated by PMS #259
package dms

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

// @API RabbitMQ GET /v2/rabbitmq/{project_id}/instances/{instance_id}/vhosts/{vhost}/queues
func DataSourceDmsRabbitmqQueues() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDmsRabbitmqQueuesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vhost": {
				Type:     schema.TypeString,
				Required: true,
			},
			"queues": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auto_delete": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"durable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"dead_letter_exchange": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dead_letter_routing_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message_ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"lazy_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"messages": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"consumers": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDmsRabbitmqQueuesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)
	client, err := cfg.NewServiceClient("dmsv2", region)
	if err != nil {
		return diag.Errorf("error creating DMS client: %s", err)
	}

	results, err := getRabbitmqQueuesList(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("error generating UUID")
	}
	d.SetId(id)

	mErr := multierror.Append(nil,
		d.Set("region", region),
		d.Set("queues", results),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func getRabbitmqQueuesList(client *golangsdk.ServiceClient, d *schema.ResourceData) (interface{}, error) {
	listHttpUrl := "v2/rabbitmq/{project_id}/instances/{instance_id}/vhosts/{vhost}/queues"
	listPath := client.Endpoint + listHttpUrl
	listPath = strings.ReplaceAll(listPath, "{project_id}", client.ProjectID)
	listPath = strings.ReplaceAll(listPath, "{instance_id}", d.Get("instance_id").(string))
	listPath = strings.ReplaceAll(listPath, "{vhost}", d.Get("vhost").(string))
	listOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}

	// pageLimit is `10`
	listPath += fmt.Sprintf("?limit=%d", pageLimit)
	offset := 0
	results := make([]map[string]interface{}, 0)
	for {
		currentPath := listPath + fmt.Sprintf("&offset=%d", offset)
		listResp, err := client.Request("GET", currentPath, &listOpt)
		if err != nil {
			return nil, fmt.Errorf("error retrieving the queues list: %s", err)
		}
		listRespBody, err := utils.FlattenResponse(listResp)
		if err != nil {
			return nil, fmt.Errorf("error flattening the queues list: %s", err)
		}

		queues := utils.PathSearch("items", listRespBody, make([]interface{}, 0)).([]interface{})
		for _, queue := range queues {
			auguments := utils.PathSearch("arguments", queue, nil)
			results = append(results, map[string]interface{}{
				"name":                    utils.PathSearch("name", queue, nil),
				"auto_delete":             utils.PathSearch("auto_delete", queue, nil),
				"durable":                 utils.PathSearch("durable", queue, nil),
				"dead_letter_exchange":    utils.PathSearch(`"x-dead-letter-exchange"`, auguments, nil),
				"dead_letter_routing_key": utils.PathSearch(`"x-dead-letter-routing-key"`, auguments, nil),
				"message_ttl":             utils.PathSearch(`"x-message-ttl"`, auguments, nil),
				"lazy_mode":               utils.PathSearch(`"x-queue-mode"`, auguments, nil),
				"messages":                utils.PathSearch("messages", queue, nil),
				"consumers":               utils.PathSearch("consumers", queue, nil),
				"policy":                  utils.PathSearch("policy", queue, nil),
			})
		}

		// `total` means the number of all `queues`, and type is float64.
		offset += pageLimit
		total := utils.PathSearch("total", listRespBody, float64(0))
		if int(total.(float64)) <= offset {
			break
		}
	}

	return results, nil
}