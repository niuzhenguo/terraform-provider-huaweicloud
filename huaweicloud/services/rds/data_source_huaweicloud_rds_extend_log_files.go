// Generated by PMS #216
package rds

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/httphelper"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/schemas"
)

func DataSourceRdsExtendLogFiles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRdsExtendLogFilesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the ID of the instance.`,
			},
			"files": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `Indicates the list of extend log files.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Indicates the file Name.`,
						},
						"file_size": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Indicates the file size. Unit: KB.`,
						},
					},
				},
			},
		},
	}
}

type ExtendLogFilesDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newExtendLogFilesDSWrapper(d *schema.ResourceData, meta interface{}) *ExtendLogFilesDSWrapper {
	return &ExtendLogFilesDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceRdsExtendLogFilesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newExtendLogFilesDSWrapper(d, meta)
	listXellogFilesRst, err := wrapper.ListXellogFiles()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listXellogFilesToSchema(listXellogFilesRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API RDS GET /v3/{project_id}/instances/{instance_id}/xellog-files
func (w *ExtendLogFilesDSWrapper) ListXellogFiles() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "rds")
	if err != nil {
		return nil, err
	}

	uri := "/v3/{project_id}/instances/{instance_id}/xellog-files"
	uri = strings.ReplaceAll(uri, "{instance_id}", w.Get("instance_id").(string))
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		OffsetPager("list", "offset", "limit", 2).
		Request().
		Result()
}

func (w *ExtendLogFilesDSWrapper) listXellogFilesToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("files", schemas.SliceToList(body.Get("list"),
			func(files gjson.Result) any {
				return map[string]any{
					"file_name": files.Get("file_name").Value(),
					"file_size": files.Get("file_size").Value(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}
