<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# UNSTABLE

This package is unstable. Unlike most other packages in platform-common-lib, it does *not* follow semver, and may experience many breaking changes

# Entity Reference Manager

Includes a common use case for protection entities from deletion.

There are two types of references:

- Hard reference - to declare that object cannot be removed without releasing the reference. In this case validation
callback URL to be used to validate if reference is still valid or not
- Soft reference - to inform that referenced service when object got removed, so that it can run some action (e.g.
  contact is removed, so that all tickets are remapped to the other default contact). In this case, notification callback URL is used to notify target service about referencing object removal. HTTP POST call should be invoked to inform target service about referenced object removal

##### Prerequisites

To use that use case the following requirements should be met:
1. Implement entityreference.Repo interface to store references in the target microservice

##### Usage

1. Call the constructor NewManagementUsecase() and pass all the arguments
2. Choose methods

##### Example

For working example please check out example/reference.go