---
subcategory: "Distributed Cache Service"
---

# huaweicloud\_dcs\_product

Use this data source to get the ID of an available Flexibleengine dcs product.
This is an alternative to `huaweicloud_dcs_product_v1`

## Example Usage

```hcl

data "huaweicloud_dcs_product" "product1" {
  spec_code = "dcs.single_node"
}
```

## Argument Reference

* `region` - (Optional, String) The region in which to obtain the dcs products. If omitted, the provider-level region will be used.

* `spec_code` - (Optional, String) DCS instance specification code. For details, see
[Querying Service Specifications](https://support.huaweicloud.com/en-us/api-dcs/dcs-api-0312040.html).


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Specifies a data source ID in UUID format.


