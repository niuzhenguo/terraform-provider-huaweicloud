// Generated by PMS #323
package secmaster

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
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func DataSourceSecmasterDataClasses() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecmasterDataClassesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the region in which to query the resource. If omitted, the provider-level region will be used.`,
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the workspace ID.`,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the name of the data class. Fuzzy matching is supported.`,
			},
			"business_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the business code of the data class. Fuzzy matching is supported.`,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the data class description. Fuzzy matching is supported.`,
			},
			"is_built_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies whether the data class is built in SecMaster.`,
			},
			"data_classes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `The data class list.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The ID of the data class.`,
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The name of the data class.`,
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The description of the data class.`,
						},
						"business_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The business code of the data class.`,
						},
						"is_built_in": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: `Whether the data class is built in SecMaster.`,
						},
						"type_num": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: `The quantity of sub-type data classes.`,
						},
						"subscribed_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The subscribed version of the data class.`,
						},
						"parent_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The parent data class ID.`,
						},
						"workspace_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The workspace ID.`,
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The creation time.`,
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The update time.`,
						},
						"creator_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The creator ID.`,
						},
						"modifier_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The modifier ID.`,
						},
						"creator": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The creator.`,
						},
						"modifier": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The modifier.`,
						},
						"domain_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The domain ID.`,
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The region ID.`,
						},
						"project_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The project ID.`,
						},
					},
				},
			},
		},
	}
}

type DataClassesDSWrapper struct {
	*schemas.ResourceDataWrapper
	Config *config.Config
}

func newDataClassesDSWrapper(d *schema.ResourceData, meta interface{}) *DataClassesDSWrapper {
	return &DataClassesDSWrapper{
		ResourceDataWrapper: schemas.NewSchemaWrapper(d),
		Config:              meta.(*config.Config),
	}
}

func dataSourceSecmasterDataClassesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	wrapper := newDataClassesDSWrapper(d, meta)
	listDataclassRst, err := wrapper.ListDataclass()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	err = wrapper.listDataclassToSchema(listDataclassRst)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// @API SecMaster GET /v1/{project_id}/workspaces/{workspace_id}/soc/dataclasses
func (w *DataClassesDSWrapper) ListDataclass() (*gjson.Result, error) {
	client, err := w.NewClient(w.Config, "secmaster")
	if err != nil {
		return nil, err
	}

	uri := "/v1/{project_id}/workspaces/{workspace_id}/soc/dataclasses"
	uri = strings.ReplaceAll(uri, "{workspace_id}", w.Get("workspace_id").(string))
	params := map[string]any{
		"name":          w.Get("name"),
		"business_code": w.Get("business_code"),
		"description":   w.Get("description"),
		"is_built_in":   w.Get("is_built_in"),
	}
	params = utils.RemoveNil(params)
	return httphelper.New(client).
		Method("GET").
		URI(uri).
		Query(params).
		OffsetPager("dataclass_details", "offset", "limit", 100).
		Request().
		Result()
}

func (w *DataClassesDSWrapper) listDataclassToSchema(body *gjson.Result) error {
	d := w.ResourceData
	mErr := multierror.Append(nil,
		d.Set("region", w.Config.GetRegion(w.ResourceData)),
		d.Set("data_classes", schemas.SliceToList(body.Get("dataclass_details"),
			func(dataClasses gjson.Result) any {
				return map[string]any{
					"id":                 dataClasses.Get("id").Value(),
					"name":               dataClasses.Get("name").Value(),
					"description":        dataClasses.Get("description").Value(),
					"business_code":      dataClasses.Get("business_code").Value(),
					"is_built_in":        dataClasses.Get("is_built_in").Value(),
					"type_num":           dataClasses.Get("type_num").Value(),
					"subscribed_version": dataClasses.Get("cloud_pack_version").Value(),
					"parent_id":          dataClasses.Get("parent_id").Value(),
					"workspace_id":       dataClasses.Get("workspace_id").Value(),
					"created_at":         dataClasses.Get("create_time").Value(),
					"updated_at":         dataClasses.Get("update_time").Value(),
					"creator_id":         dataClasses.Get("creator_id").Value(),
					"modifier_id":        dataClasses.Get("modifier_id").Value(),
					"creator":            dataClasses.Get("creator_name").Value(),
					"modifier":           dataClasses.Get("modifier_name").Value(),
					"domain_id":          dataClasses.Get("domain_id").Value(),
					"region_id":          dataClasses.Get("region_id").Value(),
					"project_id":         dataClasses.Get("project_id").Value(),
				}
			},
		)),
	)
	return mErr.ErrorOrNil()
}