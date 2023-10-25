# Packer Basic Example

This folder contains a very simple Packer template to demonstrate how you can use Terratest to write automated tests
for your Packer templates. The template just creates an up-to-date Ubuntu AMI by running `apt-get update` and
`apt-get upgrade`.

Check out [test/packer_basic_example_test.go](https://github.com/terraform-modules-krish/terratest/blob/v0.10.4/test/packer_basic_example_test.go) to see how you can write
automated tests for this simple template.

Note that this template doesn't do anything useful; it's just here to demonstrate the simplest usage pattern for
Terratest. For slightly more complicated, real-world examples of Packer templates and the corresponding tests, see
[packer-docker-example](https://github.com/terraform-modules-krish/terratest/blob/v0.10.4/examples/packer-docker-example) and
[terraform-packer-example](https://github.com/terraform-modules-krish/terratest/blob/v0.10.4/examples/terraform-packer-example).




## Building the Packer template manually

1. Sign up for [AWS](https://aws.amazon.com/).
1. Configure your AWS credentials using one of the [supported methods for AWS CLI
   tools](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html), such as setting the
   `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.
1. Install [Packer](https://www.packer.io/) and make sure it's on your `PATH`.
1. Run `packer build build.json`.




## Running automated tests against this Packer template

1. Sign up for [AWS](https://aws.amazon.com/).
1. Configure your AWS credentials using one of the [supported methods for AWS CLI
   tools](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html), such as setting the
   `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.
1. Install [Packer](https://www.packer.io/) and make sure it's on your `PATH`.
1. Install [Golang](https://golang.org/) and make sure this code is checked out into your `GOPATH`.
1. `cd test`
1. `dep ensure`
1. `go test -v -run TestPackerBasicExample`