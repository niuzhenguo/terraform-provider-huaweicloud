package apig

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance/common"
)

func TestAccDataSourceApplicationQuotas_basic(t *testing.T) {
	var (
		rName = "data.huaweicloud_apig_application_quotas.test"
		dc    = acceptance.InitDataSourceCheck(rName)

		byId   = "data.huaweicloud_apig_application_quotas.filter_by_id"
		dcById = acceptance.InitDataSourceCheck(byId)

		byName   = "data.huaweicloud_apig_application_quotas.filter_by_name"
		dcByName = acceptance.InitDataSourceCheck(byName)

		byNotFoundName   = "data.huaweicloud_apig_application_quotas.filter_by_not_found_name"
		dcByNotFoundName = acceptance.InitDataSourceCheck(byNotFoundName)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceApplicationQuotas_basic(),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestMatchResourceAttr(rName, "quotas.#", regexp.MustCompile(`^[1-9]([0-9]*)?$`)),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.name"),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.description"),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.call_limits"),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.time_unit"),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.time_interval"),
					resource.TestCheckResourceAttrSet(rName, "quotas.0.bound_app_num"),
					resource.TestMatchResourceAttr(rName, "quotas.0.created_at",
						regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}?(Z|([+-]\d{2}:\d{2}))$`)),
					dcById.CheckResourceExists(),
					resource.TestCheckOutput("is_id_filter_useful", "true"),
					dcByName.CheckResourceExists(),
					resource.TestCheckOutput("is_name_filter_useful", "true"),
					dcByNotFoundName.CheckResourceExists(),
					resource.TestCheckOutput("is_name_not_found_filter_useful", "true"),
				),
			},
		},
	})
}

func testAccDataSourceApplicationQuotas_basic_base() string {
	name := acceptance.RandomAccResourceName()

	return fmt.Sprintf(`
data "huaweicloud_availability_zones" "test" {}

%[1]s

resource "huaweicloud_apig_instance" "test" {
  name                  = "%[2]s"
  edition               = "BASIC"
  vpc_id                = huaweicloud_vpc.test.id
  subnet_id             = huaweicloud_vpc_subnet.test.id
  security_group_id     = huaweicloud_networking_secgroup.test.id
  enterprise_project_id = "0"

  availability_zones = try(slice(data.huaweicloud_availability_zones.test.names, 0, 1), null)
}

resource "huaweicloud_apig_application_quota" "test" {
  instance_id   = huaweicloud_apig_instance.test.id
  name          = "%[2]s"
  time_unit     = "MINUTE"
  call_limits   = 20
  time_interval = 2
  description   = "Created by terraform script"
}
`, common.TestBaseNetwork(name), name)
}

func testAccDataSourceApplicationQuotas_basic() string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_apig_application_quotas" "test" {
  depends_on = [
    huaweicloud_apig_application_quota.test,
  ]

  instance_id = huaweicloud_apig_instance.test.id
}

# Filter by ID
locals {
  quota_id = data.huaweicloud_apig_application_quotas.test.quotas[0].id
}

data "huaweicloud_apig_application_quotas" "filter_by_id" {
  instance_id = huaweicloud_apig_instance.test.id
  quota_id    = local.quota_id
}

locals {
  id_filter_result = [
    for v in data.huaweicloud_apig_application_quotas.filter_by_id.quotas[*].id : v == local.quota_id
  ]
}

output "is_id_filter_useful" {
  value = length(local.id_filter_result) > 0 && alltrue(local.id_filter_result)
}

# Filter by name
locals {
  quota_name = data.huaweicloud_apig_application_quotas.test.quotas[0].name
}

data "huaweicloud_apig_application_quotas" "filter_by_name" {
  instance_id = huaweicloud_apig_instance.test.id
  name        = local.quota_name
}

locals {
  name_filter_result = [
    for v in data.huaweicloud_apig_application_quotas.filter_by_name.quotas[*].name : v == local.quota_name
  ]
}

output "is_name_filter_useful" {
  value = length(local.name_filter_result) > 0 && alltrue(local.name_filter_result)
}

# Filter by name (not found)
locals {
  not_found_name = "not_found"
}

data "huaweicloud_apig_application_quotas" "filter_by_not_found_name" {
  instance_id =huaweicloud_apig_instance.test.id
  name        = local.not_found_name
}

output "is_name_not_found_filter_useful" {
  value = length(data.huaweicloud_apig_application_quotas.filter_by_not_found_name.quotas) < 1
}
`, testAccDataSourceApplicationQuotas_basic_base())
}
