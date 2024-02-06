package cloudwatchevent

import "time"

// Represents the top-level structure of a CloudTrail event.
type CloudTrailEvent struct {
	Version    string    `json:"version"`
	ID         string    `json:"id"`
	DetailType string    `json:"detail-type"`
	Source     string    `json:"source"`
	Account    string    `json:"account"`
	Time       time.Time `json:"time"`
	Region     string    `json:"region"`
	Detail     Detail    `json:"detail"`
}

// Detail holds information about the API call event.
type Detail struct {
	EventVersion       string            `json:"eventVersion"`
	UserIdentity       UserIdentity      `json:"userIdentity"`
	EventTime          time.Time         `json:"eventTime"`
	EventSource        string            `json:"eventSource"`
	EventName          string            `json:"eventName"`
	AwsRegion          string            `json:"awsRegion"`
	SourceIPAddress    string            `json:"sourceIPAddress"`
	UserAgent          string            `json:"userAgent"`
	RequestParameters  RequestParameters `json:"requestParameters"`
	ResponseElements   ResponseElements  `json:"responseElements"`
	RequestID          string            `json:"requestID"`
	EventID            string            `json:"eventID"`
	ReadOnly           bool              `json:"readOnly"`
	EventType          string            `json:"eventType"`
	ManagementEvent    bool              `json:"managementEvent"`
	RecipientAccountId string            `json:"recipientAccountId"`
	EventCategory      string            `json:"eventCategory"`
}

// UserIdentity describes the identity of the requester.
type UserIdentity struct {
	Type           string         `json:"type"`
	PrincipalID    string         `json:"principalId"`
	ARN            string         `json:"arn"`
	AccountId      string         `json:"accountId"`
	SessionContext SessionContext `json:"sessionContext"`
	InvokedBy      string         `json:"invokedBy"`
}

// SessionContext provides context for the session in which the request was made.
type SessionContext struct {
	SessionIssuer SessionIssuer `json:"sessionIssuer"`
	Attributes    Attributes    `json:"attributes"`
}

// SessionIssuer details about the entity that provided the session.
type SessionIssuer struct {
	Type        string `json:"type"`
	PrincipalID string `json:"principalId"`
	ARN         string `json:"arn"`
	AccountId   string `json:"accountId"`
	UserName    string `json:"userName"`
}

// Attributes holds session attributes such as creation time and MFA authentication status.
type Attributes struct {
	CreationDate     time.Time `json:"creationDate"`
	MfaAuthenticated string    `json:"mfaAuthenticated"`
}

// RequestParameters includes parameters specific to the API call.
type RequestParameters struct {
	SubnetId              string      `json:"subnetId"`
	Description           string      `json:"description"`
	GroupSet              GroupSet    `json:"groupSet"`
	PrivateIpAddressesSet interface{} `json:"privateIpAddressesSet"` // Can be empty, adjust based on actual use
	Ipv6AddressCount      int         `json:"ipv6AddressCount"`
	ClientToken           string      `json:"clientToken"`
}

// ResponseElements contains the response from the API call.
type ResponseElements struct {
	RequestId        string           `json:"requestId"`
	NetworkInterface NetworkInterface `json:"networkInterface"`
}

// NetworkInterface details about the created or modified network interface.
type NetworkInterface struct {
	NetworkInterfaceId    string                `json:"networkInterfaceId"`
	SubnetId              string                `json:"subnetId"`
	VpcId                 string                `json:"vpcId"`
	AvailabilityZone      string                `json:"availabilityZone"`
	Description           string                `json:"description"`
	OwnerId               string                `json:"ownerId"`
	RequesterId           string                `json:"requesterId"`
	RequesterManaged      bool                  `json:"requesterManaged"`
	Status                string                `json:"status"`
	MacAddress            string                `json:"macAddress"`
	PrivateIpAddress      string                `json:"privateIpAddress"`
	PrivateDnsName        string                `json:"privateDnsName"`
	SourceDestCheck       bool                  `json:"sourceDestCheck"`
	InterfaceType         string                `json:"interfaceType"`
	GroupSet              GroupSet              `json:"groupSet"`
	PrivateIpAddressesSet PrivateIpAddressesSet `json:"privateIpAddressesSet"`
	Ipv6AddressesSet      interface{}           `json:"ipv6AddressesSet"`
	TagSet                interface{}           `json:"tagSet"`
}

// GroupSet represents a set of security groups.
type GroupSet struct {
	Items []GroupItem `json:"items"`
}

// GroupItem details of a single security group.
type GroupItem struct {
	GroupId   string `json:"groupId"`
	GroupName string `json:"groupName"` // Not present in your JSON, included for completeness
}

// PrivateIpAddressesSet structure to match the JSON structure for private IP addresses.
type PrivateIpAddressesSet struct {
	Item []PrivateIpAddressSetItem `json:"item"`
}

// PrivateIpAddressSetItem contains details about a single private IP address.
type PrivateIpAddressSetItem struct {
	PrivateIpAddress string `json:"privateIpAddress"`
	PrivateDnsName   string `json:"privateDnsName"`
	Primary          bool   `json:"primary"`
}
