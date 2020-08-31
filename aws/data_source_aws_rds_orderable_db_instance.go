package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRdsOrderableDbInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRdsOrderableDbInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zone_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"db_instance_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"engine": {
				Type:     schema.TypeString,
				Required: true,
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"max_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"max_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"max_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"min_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"multi_az_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"outpost_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"preferred_db_instance_classes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"read_replica_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"storage_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"supported_engine_modes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supports_enhanced_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_global_databases": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_iam_database_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_iops": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_kerberos_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_performance_insights": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_storage_autoscaling": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_storage_encryption": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRdsOrderableDbInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{}

	if v, ok := d.GetOk("availability_zone_group"); ok {
		input.AvailabilityZoneGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_instance_class"); ok {
		input.DBInstanceClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_model"); ok {
		input.LicenseModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Reading RDS Orderable DB Instance Options: %v", input)
	var instanceClassResults []string
	var instanceEngineVersions []string
	instanceInfo := make(map[string]interface{})

	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *rds.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			if v, ok := d.GetOk("storage_type"); ok {
				if aws.StringValue(instanceOption.StorageType) != v.(string) {
					continue
				}
			}

			instanceClass := aws.StringValue(instanceOption.DBInstanceClass)
			instanceClassResults = append(instanceClassResults, instanceClass)
			instanceInfo[instanceClass] = instanceOption

			instanceEngineVersions = append(instanceEngineVersions, aws.StringValue(instanceOption.EngineVersion))
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading RDS orderable DB instance options: %w", err)
	}

	if len(instanceClassResults) == 0 {
		return fmt.Errorf("no RDS Orderable DB Instance options found matching criteria; try different search")
	}

	// preferred classes
	var foundInstanceClass string
	if l := d.Get("preferred_db_instance_classes").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceClass, ok := elem.(string)

			if !ok {
				continue
			}

			for _, instanceClassResult := range instanceClassResults {
				if instanceClassResult == preferredInstanceClass {
					foundInstanceClass = preferredInstanceClass
					break
				}
			}

			if foundInstanceClass != "" {
				break
			}
		}
	}

	if foundInstanceClass == "" && len(instanceClassResults) > 1 {
		return fmt.Errorf("multiple RDS DB Instance Classes (%v) match the criteria; try a different search", instanceClassResults)
	}

	if foundInstanceClass == "" && len(instanceClassResults) == 1 {
		foundInstanceClass = instanceClassResults[0]
	}

	if foundInstanceClass == "" {
		return fmt.Errorf("no RDS DB Instance Classes match the criteria; try a different search")
	}

	d.SetId(foundInstanceClass)

	d.Set("db_instance_class", foundInstanceClass)

	instanceOption := *instanceInfo[foundInstanceClass].(*rds.OrderableDBInstanceOption)

	d.Set("availability_zone_group", instanceOption.AvailabilityZoneGroup)

	var availabilityZones []*string
	for _, az := range instanceOption.AvailabilityZones {
		//availabilityZones = append(availabilityZones, aws.StringValue(az.Name))
		availabilityZones = append(availabilityZones, az.Name)
	}
	d.Set("availability_zones", availabilityZones)

	d.Set("engine", instanceOption.Engine)
	d.Set("engine_version", instanceOption.EngineVersion)
	d.Set("license_model", instanceOption.LicenseModel)
	d.Set("max_iops_per_db_instance", instanceOption.MaxIopsPerDbInstance)
	d.Set("max_iops_per_gib", instanceOption.MaxIopsPerGib)
	d.Set("max_storage_size", instanceOption.MaxStorageSize)
	d.Set("min_iops_per_db_instance", instanceOption.MinIopsPerDbInstance)
	d.Set("min_iops_per_gib", instanceOption.MinIopsPerGib)
	d.Set("min_storage_size", instanceOption.MinStorageSize)
	d.Set("multi_az_capable", instanceOption.MultiAZCapable)
	d.Set("outpost_capable", instanceOption.OutpostCapable)
	d.Set("read_replica_capable", instanceOption.ReadReplicaCapable)
	d.Set("storage_type", instanceOption.StorageType)
	d.Set("supported_engine_modes", instanceOption.SupportedEngineModes)
	d.Set("supports_enhanced_monitoring", instanceOption.SupportsEnhancedMonitoring)
	d.Set("supports_global_databases", instanceOption.SupportsGlobalDatabases)
	d.Set("supports_iam_database_authentication", instanceOption.SupportsIAMDatabaseAuthentication)
	d.Set("supports_iops", instanceOption.SupportsIops)
	d.Set("supports_kerberos_authentication", instanceOption.SupportsKerberosAuthentication)
	d.Set("supports_performance_insights", instanceOption.SupportsPerformanceInsights)
	d.Set("supports_storage_autoscaling", instanceOption.SupportsStorageAutoscaling)
	d.Set("supports_storage_encryption", instanceOption.SupportsStorageEncryption)
	d.Set("vpc", instanceOption.Vpc)

	return nil
}
