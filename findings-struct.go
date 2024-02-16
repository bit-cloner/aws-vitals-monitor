package main

type AccountInformation struct {
	AccountId    string `json:"accountId"`
	AccountAlias string `json:"accountAlias"`
	RegionCode   string `json:"regionCode"`
	RegionName   string `json:"regionName"`
}

type Snapshots struct {
	TotalAnalyzed  int  `json:"totalAnalyzed"`
	PubliclyShared bool `json:"publiclyShared"`
}

type OpenPortIssue struct {
	SecurityGroupId string `json:"securityGroupId"`
	PortRange       string `json:"portRange"`
}

type ExcessivelyOpenInboundRule struct {
	Port            int    `json:"port"`
	SecurityGroupId string `json:"securityGroupId"`
}

type SecurityGroups struct {
	TotalAnalyzed               int                          `json:"totalAnalyzed"`
	OpenPortIssues              []OpenPortIssue              `json:"openPortIssues"`
	ExcessivelyOpenInboundRules []ExcessivelyOpenInboundRule `json:"excessivelyOpenInboundRules"`
}

type LambdaFunctions struct {
	TotalAnalyzed    int  `json:"totalAnalyzed"`
	OutdatedRuntimes bool `json:"outdatedRuntimes"`
}

type RDSInstances struct {
	TotalAnalyzed int `json:"totalAnalyzed"`
}

type Bucket struct {
	Name                    string            `json:"name"`
	StorageClassPercentages map[string]string `json:"storageClassPercentages"`
}

type S3Buckets struct {
	TotalBuckets                            int      `json:"totalBuckets"`
	Buckets                                 []Bucket `json:"buckets"`
	BucketsWithoutLifecyclePolicyPercentage string   `json:"bucketsWithoutLifecyclePolicyPercentage"`
}

type DynamoDb struct {
	TotalTables       int `json:"totalTables"`
	ProvisionedTables int `json:"provisionedTables"`
	OnDemandTables    int `json:"onDemandTables"`
}

type InstanceTypeDistribution struct {
	Count      int               `json:"count"`
	Percentage string            `json:"percentage"`
	Types      map[string]string `json:"types"`
}

type InstancesAnalysis struct {
	TotalInstancesCheckedForIMDv1 int      `json:"totalInstancesCheckedForIMDv1"`
	InstancesUsingIMDv1           []string `json:"instancesUsingIMDv1"`
}

type Findings struct {
	Snapshots                 Snapshots                `json:"snapshots"`
	SecurityGroups            SecurityGroups           `json:"securityGroups"`
	DefaultSecurityGroupUsage bool                     `json:"defaultSecurityGroupUsage"`
	BroadPrivateCidrRange     bool                     `json:"broadPrivateCidrRange"`
	RepositoriesFound         bool                     `json:"repositoriesFound"`
	OrphanedElasticIPs        bool                     `json:"orphanedElasticIPs"`
	OverlappingSubnets        bool                     `json:"overlappingSubnets"`
	OrphanedEBSVolumes        bool                     `json:"orphanedEBSVolumes"`
	LambdaFunctions           LambdaFunctions          `json:"lambdaFunctions"`
	RDSInstances              RDSInstances             `json:"rdsInstances"`
	InstancesAnalysis         InstancesAnalysis        `json:"instancesAnalysis"`
	ReservedInstancePurchases int                      `json:"reservedInstancePurchases"`
	InstanceTypeDistribution  InstanceTypeDistribution `json:"instanceTypeDistribution"`
	S3Buckets                 S3Buckets                `json:"s3Buckets"`
	DynamoDb                  DynamoDb                 `json:"dynamoDb"`
}

type MasterStructure struct {
	AccountInformation AccountInformation `json:"accountInformation"`
	Findings           Findings           `json:"findings"`
}
