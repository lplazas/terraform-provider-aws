package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccDataSourceAwsDxLocation_basic(t *testing.T) {
	dsResourceName := "data.aws_dx_location.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDxLocationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsResourceName, "available_port_speeds.#"),
					resource.TestCheckResourceAttrSet(dsResourceName, "available_providers.#"),
					resource.TestCheckResourceAttrSet(dsResourceName, "location_code"),
					resource.TestCheckResourceAttrSet(dsResourceName, "location_name"),
				),
			},
		},
	})
}

const testAccDataSourceDxLocationConfig_basic = `
data "aws_dx_locations" "test" {}

data "aws_dx_location" "test" {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[0]
}
`