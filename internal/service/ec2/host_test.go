package ec2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func init() {
	resource.AddTestSweepers("aws_ec2_host", &resource.Sweeper{
		Name: "aws_ec2_host",
		F:    testSweepEc2Hosts,
		Dependencies: []string{
			"aws_instance",
		},
	})
}

func testSweepEc2Hosts(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeHostsInput{}
	sweepResources := make([]*acctest.SweepResource, 0)

	err = conn.DescribeHostsPages(input, func(page *ec2.DescribeHostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, host := range page.Hosts {
			r := ResourceHost()
			d := r.Data(nil)
			d.SetId(aws.StringValue(host.HostId))

			sweepResources = append(sweepResources, acctest.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if acctest.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Host sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Hosts (%s): %w", region, err)
	}

	err = acctest.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Hosts (%s): %w", region, err)
	}

	return nil
}

func TestAccAWSEc2Host_basic(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "a1.large"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2Host_disappears(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceHost(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2Host_InstanceFamily(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfigInstanceFamily(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "off"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "on"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "c5"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2HostConfigInstanceType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dedicated-host/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "off"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "c5.xlarge"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSEc2Host_Tags(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_ec2_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEc2HostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2HostConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2HostConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2HostConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2HostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEc2HostExists(n string, v *ec2.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Host ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindHostByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEc2HostDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_host" {
			continue
		}

		_, err := tfec2.FindHostByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Host %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSEc2HostConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"
}
`)
}

func testAccAWSEc2HostConfigInstanceFamily(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "off"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "on"
  instance_family   = "c5"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSEc2HostConfigInstanceType(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "on"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "off"
  instance_type     = "c5.xlarge"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSEc2HostConfigTags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAWSEc2HostConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}