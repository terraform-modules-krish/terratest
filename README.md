***WARNING: THIS REPO IS AN AUTO-GENERATED COPY.*** *This repo has been copied from [Gruntwork’s](https://gruntwork.io/) GitHub repositories so that you can consume it from your company’s own internal Git repositories. This copy is automatically created and updated by the `repo-copier` CLI tool. If you need to make changes to this repo, you should make the changes in a separate fork, and NOT make changes directly in this repo, as otherwise, the `repo-copier` will overwrite your changes! Please see the `repo-copier` [documentation](https://github.com/terraform-modules-krish/repo-copier) for more information on how the code is copied, how cross-references are updated, how the changelog is handled, etc.*

***

_You may find it valuable to view the following resources in the original repo. If these links give you a 404, visit https://app.gruntwork.io to gain access or email support@gruntwork.io if you need assistance._

[Home Page](https://github.com/gruntwork-io/terratest/) |
[Pull Requests](https://github.com/gruntwork-io/terratest/pulls) |
[Issues](https://github.com/gruntwork-io/terratest/issues) |
[Releases and Assets](https://github.com/gruntwork-io/terratest/releases)

_Alternatively, you can view a copied version of the resources listed above._

[Pull Requests](https://github.com/terraform-modules-krish/terratest/blob/master/.github/PULL_REQUESTS.md) |
[Issues](https://github.com/terraform-modules-krish/terratest/blob/master/.github/ISSUES.md) |
[ChangeLog](https://github.com/terraform-modules-krish/terratest/blob/master/.github/CHANGELOG.md)

***

# Terraform Test Framework

This framework runs unit tests and integration tests for terraform modules.



## The Big Idea

Gruntwork customers rely on our Terraform Modules to build their infrastructure, so we need a high degree of confidence that our modules work as expected.

What would give us such confidence?

- The module should correctly `terraform apply` and `terraform destroy`.
- The module should not become "broken" as other updates are made to it or other modules.
- Each module should work in every AWS region.
- Once the module's infrastructure is created, functionality should work as expected.

The goal of this framework is to automate everything on the list above.  We want to make this automation fast by parallelizing tests, and we want to make it
easy to write tests in the first place.

### Traditional Software Tests vs. DevOps Software Tests

In traditional software testing, we have the benefit of near total-control over our local environment.  If our code relies on third-party services, we can "mock"
these services to simulate their responses so that our laptops can very quickly validate that our code works as expected.

In DevOps, we can't mock AWS.  It's simply too complex with too many interacting services.  So we accept this as a limitation of our environment, and take
steps to introduce other properties of unit tests such as parallelizability, isolation, and fastness.

### Achieving Unit Tests for Our Terraform Modules

We approximate unit tests for our Terraform Modules with the following philosophy:

- Every AWS resource Terraform creates should be "namespaced" so that the same template running a unit test in parallel won't have a resource name conflict.
- We should randomize the AWS region in which we launch our resources to expose resource-specific quirks such different numbers of Availability Zones.
- We should auto-generate select resources such as EC2 KeyPairs since many Terraform Modules assume these are created "out of band".

By adopting this philosophy, terraform-test can run multiple "unit" tests in parallel, reducing our build times!



## Warnings & Notes

**Note #1**: All of these tests, including the unit tests, create real resources in an AWS account. That means they cost money to run, especially if you don't 
clean up after yourself. Please be considerate of the resources you create and take extra care to clean everything up when you're done!

**Note #2**: Never hit `CTRL + C` or cancel a build once tests are running or the cleanup tasks won't run!

**Note #3**: We set `-timeout 30m` on all tests not because they necessarily take 30 minutes, but because Go has a default test timeout of 10 minutes, 
after which it does a `SIGQUIT`, preventing the tests from properly cleaning up after themselves. Therefore, we set a timeout of 30 minutes to make
 sure all tests have enough time to finish and cleanup.



## Concepts
### Types of Tests

We have two types of tests:

1. **Unit tests**: These are meant to be used for rapid feedback during development. They should run quickly, be able to run in parallel with other unit tests,
   and verify the functionality of a single module in total isolation.
2. **Integration tests**: These are meant to verify that many different Terraform Modules work together correctly.  In particular, we want to verify the
   examples listed in the in the [terraform-modules/examples folder](https://github.com/gruntwork-io/terraform-modules/tree/master/examples) by running a real-world, end-to-end test. Integration tests will take longer to run, 
   so we usually only run them in a CI job as a sanity check before publishing a new release.



## Usage

### Prerequisites

1. Install [Go](https://golang.org/).
2. Install [Terraform](https://www.terraform.io/downloads.html)
3. Setup the environment variables required by the unit test you want to run. The most common ones are:
   `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `TF_VAR_aws_region`. Note that your AWS user must have the right
   IAM permissions for the test.
4. Some of the tests use Terraform remote state, so you need to create an S3 bucket in the region `TF_VAR_aws_region`
   and set the bucket's name as an environment variable called `REMOTE_STATE_BUCKET`.

### Writing your unit test

We use Go's built-in [package testing](https://golang.org/pkg/testing/) for tests.  Therefore, simply create a file ending in `_test.go` and write your test.  

For samples, see the [terratest_test.go](_terratest_test.go) file which shows many examples of how to use terratest.

### Running unit tests

To run a specific unit test:

```bash
cd /location/of/go/file/with/unit/test/code.go
go test -timeout 30m -parallel 32
```

To run all unit tests, note that we do not currently have a go-native way to run all unit tests, so this requires something out-of-band like a bash script.

### Running integration tests

Integration tests run just like unit tests but may have the function signature `func TestIntegrationXXX(t *testing.T)`.   To run them:

```bash
cd /location/of/go/file/with/integration/tests.go
go test -run Integration -timeout 30m -parallel 32
```

The `-run Integration` portion tells go to run all tests with the word `Integration` in the function name.

### Running all tests

```bash
cd test
go test -timeout 30m -parallel 32
```



## Best Practices

Testing infrastructure as code (IaC) is hard. With general purpose programming languages (e.g., Java, Python, Ruby), 
you have a "localhost" environment where you can run and test the code before you commit. You can also isolate parts
of your code from external dependencies to create fast, reliable unit tests. With IaC, neither of these advantages is
available, as there isn't a "localhost" equivalent for most IaC code (e.g., I can't use Terraform to deploy an AWS
VPC on my own laptop) and there's no way to isolate your code from the outside world (i.e., the whole point of 
Terraform is to make calls to AWS, so if you remove those, there's nothing left).

That means that most of the tests are going to be integration tests that deploy into a real AWS account. This makes
the tests effective at catching real-world bugs, but it also makes them much slower and more brittle. In this section,
we'll outline some best practices to minimize the downsides of this sort of testing.  

### Unit tests

It's nearly impossible to create a truly isolated "unit" with IaC, so when we say "unit tests," what we mean is tests
that are designed to be used in the dev environment for rapid iteration. You want to be able to make a change, run the
tests, get feedback *in seconds* (ar at most, 1-2 minutes) on whether things are working, make another change, run the 
tests again, and so on. 

In other words, it's best to think of the unit tests as a dev tool, rather than a "proof of correctness." The 
integration tests (discussed below) will hopefully catch any bugs that slip by the unit tests, so make trade-offs that
let the unit tests run faster while still catching *enough* bugs that we don't need to run the integration tests too 
often.

Some unit test best practices:

#### Script modules

It should be possible to test just about any "script module" (i.e., any module written in Bash, Python, or Go) locally 
using Docker. Whereas starting and stopping an EC2 Instance can take 2-4 minutes, Docker containers start up and shut 
down in less than a second. Similarly, creating AMIs with Packer requires 2-4 minutes of overhead of starting and 
stopping an EC2 instance, as well as another 1-2 minutes to take a snapshot of it, whereas building Docker images has 
virtually no overhead, and can typically use caching to run extremely quickly.  

Here are some techniques we use with Docker:

* If your script module is used in a Packer template, add a [Docker 
  builder](https://www.packer.io/docs/builders/docker.html) to the template so you can create a Docker image from the 
  same code. You'll want to use the [docker-tag post processor](https://www.packer.io/docs/post-processors/docker-tag.html) 
  to give the image a tag you can use to run it. Example:
  
    ```json
    {
      "builders": [{
        "name": "ubuntu-ami",
        "type": "amazon-ebs"
        // ... (other params omitted) ...     
      },{
        "name": "ubuntu-docker",
        "type": "docker",
        "image": "gruntwork/ubuntu:16.04",
        "commit": "true"
      }],
      "provisioners": [
        // ...
      ],
      "post-processors": [{
        "type": "docker-tag",
        "repository": "gruntwork/example",
        "tag": "latest",
        "only": ["ubuntu-docker"]
      }]
    }
    ```  

* There are Docker images for all the Linux distros we typically support:
  [ubuntu](https://hub.docker.com/_/ubuntu/), [amazonlinux](https://hub.docker.com/_/amazonlinux/), 
  [centos](https://hub.docker.com/_/centos/). Note that the base AMIs Amazon provides for the Linux distros don't 
  always have the exact same software installed as the Docker images, so you may need to add an extra `script` 
  provisioner to "normalize" them (use the [-only flag](https://www.packer.io/docs/commands/build.html#only-foo-bar-baz)
  to target just the Docker builders). For Packer template example above, with the Ubuntu Docker image, you will 
  probably want to add:
  
    ```json
    {
      // ... (builders omitted) ...
      "provisioners": [{
        "type": "shell",
        "inline": [
          "DEBIAN_FRONTEND=noninteractive apt-get update",
          "apt-get install -y sudo curl wget ca-certificates rsyslog"
        ],
        "only": ["ubuntu-docker"]
      }]
    }
    ```   

* Create a `docker-compose.yml` for running your Docker image at test time. This file can configure any ports and
  environment variables your code needs. 
  
* Use mocks to replace external dependencies where possible. One way to do this is to create some "mock scripts" and to 
  bind-mount them in `docker-compose.yml`. For example, if your script module calls `mount-ebs-volume` to attach and 
  mount an EBS volume, you could create a mock version of `mount-ebs-volume` that simply creates a folder using 
  `mkdir -p`, and put your mock script earlier in the `PATH`. You could similarly mock out the `aws` CLI, the EC2 
  metadata endpoint, and many other external dependencies so that your script module can be tested completely on
  `localhost`!

* When writing the test code in Go, use the naming convention `TestUnitXXX` (e.g., 
  `func TestUnitJenkinsModule(t *testing.T)`) for test functions that execute these sorts of "unit tests." This way, 
  we can execute all unit tests by running `go test -run TestUnit`. 


#### Terraform modules

There's no way to deploy Terraform code on localhost, so all the tests for Terraform modules will have roughly the same 
structure:
 
1. Setup: build any AMIs and Docker images you need (e.g., using `packer build`) and deploy your code into a real AWS 
   account (e.g., using `terraform apply`).
1. Validation: test the code running in AWS actually works the way you expect it to (e.g., make HTTP requests, SSH to
   servers, etc).
1. Teardown: undeploy the code from AWS so we don't get charged for it (e.g., using `terraform destroy` in a `defer` 
   statement).

Here's a typical test case of this sort:

```go
func TestExample(t *testing.T) {
  testPath := "../examples/foo"
  logger := terralog.NewLogger("TestExample")

  amiId := buildAmi(t)                                                               // setup
  resourceCollection := createResourceCollection(t)                                  // setup
  terratestOptions := createOptions(t, amiId, testPath, resourceCollection)          // setup
  deployInfrastructureWithTerraform(t, terratestOptions)                             // setup

  defer undeployInfrastructureWithTerraform(t, resourceCollection,terratestOptions)  // teardown

  testInfrastructureWorks(t, terratestOptions)                                       // validation
}
```

These steps—especially the setup and teardown—can take a long time (5 - 30 minutes). Having to run the whole setup and
teardown process every time you change a single line of code slows down local iteration to the point of being nearly 
unusable. There's no solution that will magically make this all 100x faster, but if we structure the test code 
correctly, we can isolate each of the test stages and minimize the number of times we have to run each one.

To make this possible, you can use the methods in Terratest's `test_structure` package to structure your test case as 
follows:

```go
func TestExample(t *testing.T) {
  testPath := "../examples/foo"
  logger := terralog.NewLogger("TestExample")

  test_structure.RunTestStage("setup", logger, func() {
    amiId := buildAmi(t)                                                               // setup
    resourceCollection := createResourceCollection(t)                                  // setup
    terratestOptions := createOptions(t, amiId, testPath, resourceCollection)          // setup
    deployInfrastructureWithTerraform(t, terratestOptions)                             // setup

    test_structure.SaveTerratestOptions(t, testPath, terratestOptions)                 // save TerratestOptions for later steps
    test_structure.SaveRandomResourceCollection(t, testPath, resourceCollection)       // save RandomResourceCollection for later steps
  })

  defer test_structure.RunTestStage("teardown", logger, func() {
    terratestOptions := test_structure.LoadTerratestOptions(t, testPath)               // load TerratestOptions from earlier setup
    resourceCollection := test_structure.LoadRandomResourceCollection(t, testPath)     // load RandomResourceCollection from earlier setup

    undeployInfrastructureWithTerraform(t, resourceCollection, terratestOptions)       // teardown

    test_structure.CleanupTerratestOptions(t, testPath)                                // clean up the stored TerratestOptions
    test_structure.CleanupRandomResourceCollection(t, testPath)                        // clean up the stored RandomResourceCollection
  })

  test_structure.RunTestStage("validation", logger, func() {
    terratestOptions := test_structure.LoadTerratestOptions(t, testPath)               // load TerratestOptions from earlier setup
    testInfrastructureWorks(t, terratestOptions)                                       // validation
  })
}
```

The main change is that we've wrapped each stage of the test in a call to `test_structure.RunTestStage`. This allows us
to skip any of the test stages using an environment variable of the format `SKIP_<stage>`! For example, if you set 
`SKIP_teardown=true`, then the test code will skip the teardown process and leave your code running in AWS. This allows 
you to run the test next time with `SKIP_setup=true` and `SKIP_teardown=true` to go straight to the validation steps 
without having to wait for setup and teardown again! 

With this approach, the typical workflow will be:

1. Do the initial setup (just once): `SKIP_validation=true SKIP_teardown=true go test -run TestExample`
1. Do your validation (as many times as you want): `SKIP_setup=true SKIP_teardown=true go test -run TestExample`
1. Do the teardown (just once): `SKIP_setup=true SKIP_validation=true go test -run TestExample`

This way, you only pay the cost of setup and teardown once and you can do as many iterations on validation in
between as you want. And since the code continues to run in your AWS account, you can manually run `terraform` to 
redeploy small parts of it whenever you want, rather than having to redeploy the entire thing every time. 

Note that since any stage can be skipped, test data that needs to be available across multiple stages (e.g., 
`TerratestOptions` and `RandomResourceCollection`) are saved to disk in the setup stage and loaded from disk in 
subsequent stages.

### Integration tests

Every module should have a set of examples in the `examples` folder, and each of these examples should have an 
integration test that (a) deploys the example into a real AWS account, (b) validates it works as expected, and (c) 
undeploys the example so AWS doesn't keep charging us for it. It turns out that these are the exact same steps we use 
for "unit testing" our Terraform modules!

In other words, all integration tests should be written using the structure shown in the 
[#Terraform modules](#terraform-modules) section. In the dev environment, you can use environment variables to execute
only certain stages from these tests so you can get them working more quickly. In the CI environment, none of the 
`SKIP_XXX` environment variables will be set, so all steps will execute from start to finish. 



## License

Please see [LICENSE.txt](https://github.com/terraform-modules-krish/terratest/blob/v0.7.13/LICENSE.txt) for details on how the code in this repo is licensed.
