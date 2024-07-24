// Generated by PMS #244
package dws

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/httphelper"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/schemas"
)

func DataSourceDwsQuotas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDwsQuotasRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"quotas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `All quotas that match the filter parameters.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The type of the quota.`,
						},
						"used": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: `The number of quotas used.`,
						},
						"limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: `The number of available quotas.`,
						},
						"unit": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The unit of the quota.`,
						},
					},
				},
			},
		},
	}
}

type QuotasDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newQuotasDSWrapper(d *schema.ResourceData, meta interface{}) *QuotasDSWrapper {
	return &QuotasDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceDwsQuotasRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newQuotasDSWrapper(d, meta)
	listQuotasRst, err := wrapper.ListQuotas()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listQuotasToSchema(listQuotasRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API DWS GET /v1.0/{project_id}/quotas
func (w *QuotasDSWrapper) ListQuotas() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "dws")
	if err != nil {
		return nil, err
	}

	uri := "/v1.0/{project_id}/quotas"
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		Request().
		Result()
}

func (w *QuotasDSWrapper) listQuotasToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("quotas", schemas.SliceToList(body.Get("quotas.resources"),
			func(quotas gjson.Result) any {
				return map[string]any{
					"type":  quotas.Get("type").Value(),
					"used":  quotas.Get("used").Value(),
					"limit": quotas.Get("quota").Value(),
					"unit":  quotas.Get("unit").String(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}