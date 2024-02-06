# tailscale-route-tiler

The `tailscale-route-tiler` is a helper tool designed to automate the management of network routes for Tailscale, using configuration changes detected through AWS CloudWatch and SQS. This utility listens for messages on a specified AWS SQS queue, triggered by CloudWatch events, indicating when route updates are needed. It requires a YAML configuration file for initial setup SQS queue details, and Tailscale authentication information.

## AWS Setup

To use `tailscale-route-tiler`, you must configure AWS CloudWatch and SQS services to trigger route updates:

1. **SQS Queue Creation**: Create an SQS queue in AWS to receive CloudWatch events. Note the queue URL and ARN for configuration.

2. **CloudWatch Rule Configuration**: Set up a CloudWatch rule to monitor specific events

```json
{
  "source": ["aws.ec2"],
  "detail-type": ["AWS API Call via CloudTrail"],
  "detail": {
    "sourceIPAddress": ["elasticloadbalancing.amazonaws.com"],
    "eventName": ["CreateNetworkInterface", "DeleteNetworkInterface"],
    "requestParameters": {
      "description": [{
        "prefix": "ELB app/xxxxxxxx/xxxxxxx"
      }, {
        "prefix": "ELB app/xxxxxxx/xxxxxx"
      }]
    }
  }
}
```

3. **IAM Permissions**: Ensure the IAM role or user running `tailscale-route-tiler` has permissions to read from the SQS queue and execute necessary Tailscale API calls.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage",
                "sqs:GetQueueAttributes"
            ],
            "Resource": "arn:aws:sqs:your-region:your-account-id:your-queue-name"
        }
    ]
}
```

### Configuration

The tool requires a YAML configuration file with the following structure:

```yaml
TailscaleclientId: "EXAMPLE-CLIENT-ID"
TailscaleKey: "DUMMY-KEY"
subnets:
  - 10.49.0.0/24
  - 10.0.0.22/32
  - 10.0.0.0/24
sites:
  - google.com
TailscaleCommand: /usr/bin/tailscale up --accept-dns=false --advertise-routes=%s
EnableIpv6: false
Slack:
  WebhookURL: https://hooks.slack.com/services/EXAMPLE/EXAMPLE/EXAMPLE
  Enabled: true
SQS:
  QueueURL: https://sqs.us-west-2.amazonaws.com/EXAMPLE/EXAMPLE
  Region: us-west-2
```

## Usage

```bash
  tailscale-route-tiler worker -c config.yaml
```

## Help

```bash
tailscale-route-tiler -h
tailscale-route-tiler run -h
tailscale-route-tiler get-client-routes -h
```
