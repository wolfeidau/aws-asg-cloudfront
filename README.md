# aws-asg-cloudfront

This project demonstrates how to combine cloudfront, EC2, route53 and auto scaling groups to build a scalable http service with any load balancers.

**NOTE:** This project is a work in progress, I am shaving a few yaks along way.

# Why?

The reason I built this example is that application load balancers introduce an unnecessary layer of complexity and cost in cases were your already using a CDN such as cloudfront in front of your application.

# Diagram

TODO

# Links

- [Graviton2: ARM comes to Lambda](https://awsteele.com/blog/2021/09/29/graviton2-arm-comes-to-lambda.html)
- https://github.com/aws-samples/amazon-ec2-auto-scaling-group-examples
- https://github.com/meltwater/terraform-aws-asg-dns-handler

# License

This project is released under Apache 2.0 license and is copyright [Mark Wolfe](https://www.wolfe.id.au).
