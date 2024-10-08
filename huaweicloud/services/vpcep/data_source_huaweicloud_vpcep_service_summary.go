// Generated by PMS #320
package vpcep

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/httphelper"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/schemas"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func DataSourceVpcepServiceSummary() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVpcepServiceSummaryRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"endpoint_service_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the ID of the VPC endpoint service.`,
			},
			"endpoint_service_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the name of the VPC endpoint service.`,
			},
			"service_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The name of the VPC endpoint service.`,
			},
			"service_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The type of the VPC endpoint service.`,
			},
			"is_charge": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: `Whether the VPC endpoint connected to the VPC endpoint service is charged.`,
			},
			"enable_policy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: `Whether the VPC endpoint policy can be customized.`,
			},
			"public_border_group": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The public border group information about the pool corresponding to the VPC endpoint.`,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The creation time of the VPC endpoint service, in RFC3339 format.`,
			},
		},
	}
}

type ServiceSummaryDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newServiceSummaryDSWrapper(d *schema.ResourceData, meta interface{}) *ServiceSummaryDSWrapper {
	return &ServiceSummaryDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceVpcepServiceSummaryRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newServiceSummaryDSWrapper(d, meta)
	lisSerDesDetRst, err := wrapper.ListServiceDescribeDetails()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(lisSerDesDetRst.Get("id").String())

	err = wrapper.listServiceDescribeDetailsToSchema(lisSerDesDetRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API VPCEP GET /v1/{project_id}/vpc-endpoint-services/describe
func (w *ServiceSummaryDSWrapper) ListServiceDescribeDetails() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "vpcep")
	if err != nil {
		return nil, err
	}

	uri := "/v1/{project_id}/vpc-endpoint-services/describe"
	params := map[string]any{
		"endpoint_service_name": w.Get("endpoint_service_name"),
		"id":                    w.Get("endpoint_service_id"),
	}
	params = utils.RemoveNil(params)
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		Query(params).
		Request().
		Result()
}

func (w *ServiceSummaryDSWrapper) listServiceDescribeDetailsToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("service_name", body.Get("service_name").Value()),
		d.Set("service_type", body.Get("service_type").Value()),
		d.Set("is_charge", body.Get("is_charge").Value()),
		d.Set("enable_policy", body.Get("enable_policy").Value()),
		d.Set("public_border_group", body.Get("public_border_group").Value()),
		d.Set("created_at", w.setCreatedAt(*body)),
	)
	return mErr.ErrorOrNil()
}

func (*ServiceSummaryDSWrapper) setCreatedAt(body gjson.Result) string {
	return utils.FormatTimeStampRFC3339(utils.ConvertTimeStrToNanoTimestamp(body.Get("created_at").String())/1000, false)
}
