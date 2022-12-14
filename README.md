# aws-asg-cloudfront

This project demonstrates how to combine cloudfront, EC2, route53 and auto scaling groups to build a scalable http service with any load balancers.

**NOTE:** This project is a work in progress, I am shaving a few yaks along way.

# Why?

The reason I built this example is that application load balancers introduce an unnecessary layer of complexity and cost in cases were your already using a CDN such as cloudfront in front of your application.

# Setup

To deploy this solution you need to set some environment variables.

```
export HOSTED_ZONE_NAME=example.com
export HOSTED_ZONE_ID=ZZZZZZZZZZZZZZ
export DESIRED_CAPACITY=2
# Unique identifier used to build hostname
export INTERNAL_SERVICE_ID=aaaaaaaaaaaa
```

# Diagram

TODO

# Links

- [Graviton2: ARM comes to Lambda](https://awsteele.com/blog/2021/09/29/graviton2-arm-comes-to-lambda.html)
- https://github.com/aws-samples/amazon-ec2-auto-scaling-group-examples
- https://github.com/meltwater/terraform-aws-asg-dns-handler
- https://cloudonaut.io/cloudfront-prefix-list-security-group/

# License

This project is released under Apache 2.0 license and is copyright [Mark Wolfe](https://www.wolfe.id.au).
