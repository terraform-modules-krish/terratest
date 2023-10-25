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

### Unit tests

We use unit tests for rapid feedback while coding, so they should satisfy the following constraints:

1. Each test should run quickly (preferably less than 1 minute).
2. Each test should focus on verifying the functionality of a single module.
3. Each test must be *completely* isolated from all other tests so we can run many such tests (even many copies of the
   same unit test) in parallel in the same AWS account. That means that any resource the unit test creates, such as an
   EC2 instance or IAM role, must have a globally unique name.

For each module `XXX`, you should:

1. Have a suite of unit tests defined in a common location in your repo with subfolders for each Terraform Module.
2. The test functions within the test suite should have the following structure:

   ```go
   func TestUnitYYY(t *testing.T) {
       // Optional since you may prefer a tool like GNU Parallel to sequence output in a more sane way.
       t.Parallel()
   }
   ```
   Where `YYY` explains what the test does, such as `TestUnitAlarmNotifications`. Note that the first line of the test
   calls `t.Parallel()` to tell Go to execute multiple unit tests in parallel.
3. Have a minimal terraform template that consumes your Terraform Module that may or may not reflect real-world usage.  The goal is to test your one module in isolation.  For tests between modules, we use integration tests.

### Integration tests

We recommmend one large integration test suite that does an end-to-end test of all your Terraform Modules.  It ensures that all modules work together correctly and tries to verify them in a more real-world setting than unit testing. As a result, the integration test takes quite a bit longer to run (~20+ minutes) 
and is mostly used as a sanity check before we put out a new release of terraform-modules.

### CI

Setup a build system like [Circle CI](circleci.com) to run tests after every commit. We recommend the following protocol:

- On the `master` branch, every commit runs ALL unit tests and integration tests. 
- On all other branches, no tests are run by default.
- When a PR is submitted, all unit tests are run.



## ToDo
1. Add `circle.yml` to run tests automatically on each commit to `master`.
2. Add the following automated tests:
   1. SSH via bastion jump host to an EC2 instance.
   2. SSH directly to EC2 instance (should fail).
   3. Check exposed ports of servers in the ASG example.
   4. Take server down in the ASG example to check the ASG brings it back up.
3. Add a script that can scrub an AWS account and clean up anything left behind by accident after all the tests have
   completed. This can be done by tagging all resources created with a `terraform-module-test` tag and running a script
   that finds all resources with that tag and deletes them.

## License

Please see [LICENSE.txt](https://github.com/terraform-modules-krish/terratest/blob/v0.2.4/LICENSE.txt) for details on how the code in this repo is licensed.
