package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRdsOrderableDbInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	engine := "mysql"
	engineVersion := "5.7.22"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfigBasic(engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	engine := "mysql"
	engineVersion := "5.7.22"
	preferredOption := "db.t2.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfigPreferred(engine, engineVersion, preferredOption),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", preferredOption),
				),
			},
		},
	})
}

func testAccPreCheckAWSRdsOrderableDbInstance(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine: aws.String("mysql"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfigBasic(engine, version string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  db_instance_class = "db.t2.small"
  engine            = %q
  engine_version    = %q
  license_model     = "general-public-license"
  storage_type      = "standard"
}
`, engine, version)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfigPreferred(engine, version, preferredOption string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = %q
  engine_version = %q
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_db_instance_classes = ["db.xyz.xlarge", %q, "db.t3.small"]
}
`, engine, version, preferredOption)
}
