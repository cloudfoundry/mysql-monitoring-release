# Multierror

`multierror` is a simple go package that allows you to combine and present multiple errors as a single error.

# How to use

```go
import "github.com/cloudfoundry/multierror"

errors := multierror.MultiError{}

err1 := FirstFuncThatReturnsError()
err2 := SecondFuncThatReturnsError()

errors.Add(err1)
errors.Add(err2)

//You can also add multierror structs and they will be flattened into one struct
errors2 := multierror.MultiError()
errors2.Add(err1)
errors.Add(errors2)


//Returns the errors as an aggregate of all the error messages
errors.Error()
```

## Development

### <a name="dependencies"></a>Dependencies

This repository's dependencies are managed using
[routing-release](https://github.com/cloudfoundry/routing-release). Please refer to documentation in that repository for setting up tests

### Executables

1. `bin/test.bash`: This file is used to run test in Docker & CI. Please refer to [Dependencies](#dependencies) for setting up tests.

### Reporting issues and requesting features

Please report all issues and feature requests in [cloudfoundry/routing-release](https://github.com/cloudfoundry/routing-release).
