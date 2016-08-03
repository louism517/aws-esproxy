# aws-esproxy

Yes, other AWS Elasticsearch proxies do exist.

But this one provides the ability to use assumed STS credentials, and will will auto-refresh them upon expiry.

## Usage

```bash
# aws-esproxy -h
Usage of ./reverse-proxy:
  -arn="": ARN of role to be assumed. If ommitted then the default credential chain is used instead.
  -debug=false: Run in Debug mode.
  -esurl="": URL of AWS Elasticsearch endpoint
  -port="8080": Port to serve proxy on.
  -region="us-east-1": AWS region to use when assuming STS role.
  -session="": Session name to be used when assuming STS role.
```

The URL of the Elasticsearch endpoint must be provided (`esurl`).

If the `arn` is omitted, the proxy will find credentials using the default AWS credential chain and use them to sign requests.

Otherwise the proxy will find credentials using the default AWS credential chain and use those to assume the role specified by `arn`. It will use the `session` argument as a session name.

## Example IAM Configuration

Given an STS role ROLE and a session name SESSION, the access policy on your AWS Elasticsearch Domain should look like this:

```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:sts::123456789111:assumed-role/ROLE/SESSION"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:REGION:123456789111:domain/DOMAIN/*"
    },
  ]
}
```

The IAM credentials found by the proxy (i.e. the credentials in the surrounding enviroment, be they environment variables or an instance profile) must be allowed to assume the role ROLE, with a policy something like:

```
{
    "Statement": {
        "Resource": [
            "arn:aws:iam::123456789111:role/ROLE"
        ],
        "Effect": "Allow",
        "Action": [
            "sts:AssumeRole"
        ]
    }
}
```

Credential setups such as this are most often used for cross account access, but do have other applications.

## Development

Package vendoring is performed using `govendor`.

Pull this repo and run `govendor sync` to pull down requirements.

If you are using Go 1.5 set the environment variable `GO15VENDOREXPERIMENT=1`

## Contributing

Pull requests are welcomed.
