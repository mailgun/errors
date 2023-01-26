# Errors
A modern error handling package to add additional structured fields to errors. This allows you to keep the
[only handle errors once rule](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) while not losing context where the error occurred.

* `errors.Wrap(err, "while reading")` includes a stack trace so logging can report the exact location where
  the error occurred. *You can also call `Wrapf()`*
* `errors.WithStack(err)` for when you don't need a message, just a stack trace to where the error occurred.
* `errors.WithFields{"fileName": fileName}.Wrap(err, "while reading")` Attach additional fields to the error and a stack
  trace to give structured logging as much context to the error as possible. *You can also call `Wrapf()`*
* `errors.WithFields{"fileName": fileName}.WithStack(err)` for when you don't need a message, just a stack
  trace and some fields attached.
* `errors.WithFields{"fileName": fileName}.Error("while reading")` when you want to create a string error with
  some fields attached. *You can also call `Errorf()`*

### Extract structured data from wrapped errors
Convenience functions to extract all stack and field information from the error.
* `errors.ToLogrus() logrus.Fields`
* `errors.ToMap() map[string]interface{}`

### Example
```go
err := io.EOF
err = errors.WithFields{"fileName": "file.txt"}.Wrap(err, "while reading")
m := errors.ToMap(err)
fmt.Printf("%#v\n", m)
// OUTPUT
// map[string]interface {}{
//   "excFileName":"/path/to/wrap_test.go",
//   "excFuncName":"my_package.ReadAFile",
//   "excLineNum":42,
//   "excType":"*errors.errorString",
//   "excValue":"while reading: EOF",
//   "fileName":"file.txt"
//  }
```

## Convenience to std error library methods
Provides pass through access to the standard `errors.Is()`, `errors.As()`, `errors.Unwrap()` so you don't need to
import this package and the standard error package.

## Supported by internal tooling
If you are working at mailgun and are using scaffold; using `logrus.WithError(err)` will cause logrus to 
automatically retrieve the fields attached to the error and index them into our logging system as separate
searchable fields.

## Perfect for passing additional information to http handler middleware
If you have custom http middleware for handling unhandled errors, this is an excellent way
to easily pass additional information about the request up to the error handling middleware.

## Adding structured fields to an error
Wraps the original error while providing structured field data
```go
_, err := ioutil.ReadFile(fileName)
if err != nil {
        return errors.WithFields{"file": fileName}.Wrap(err, "while reading")
}
```

## Retrieving the structured fields
Using `errors.WithFields{}` stores the provided fields for later retrieval by upstream code or structured logging
systems
```go
// Pass to logrus as structured logging
logrus.WithFields(errors.ToLogrus(err)).Error("open file error")
```

## Support for standard golang introspection functions
Errors wrapped with `errors.WithFields{}` are compatible with standard library introspection functions `errors.Unwrap()`,
`errors.Is()` and `errors.As()`
```go
ErrQuery := errors.New("query error")
wrap := errors.WithFields{"key1": "value1"}.Wrap(err, "message")
errors.Is(wrap, ErrQuery) // == true
```

## Proper Usage
The fields wrapped by `errors.WithFields{}` are not intended to be used to by code to decide how an error should be 
handled. It is intended as a convenience where the failure is well known, but the context is dynamic. In other words,
you know the database returned an unrecoverable query error, but you want to attach localized context information
to the error.

As an example
```go
func (r *Repository) FetchAuthor(customerID, isbn string) (Author, error) {
    // Returns ErrorNotFound{} if not exist
    book, err := r.fetchBook(isbn)
    if err != nil {
        return nil, errors.WithFields{"customer.id": customerID, "isbn": isbn}.Wrap(err, "while fetching book")
    }
    // Returns ErrorNotFound{} if not exist
    author, err := r.fetchAuthorByBook(book)
    if err != nil {
        return nil, errors.WithFields{"customer.id" customerID, "book": book}.Wrap(err, "while fetching author")
    }
    return author, nil
}
```
Now you can easily search your structured logs for errors related to `customer.id`.

You should continue to create and inspect custom error types
```go

type ErrAuthorNotFound struct {
    Msg string
}

func (e *ErrAuthorNotFound) Error() string {
    return e.Msg
}

func (e *ErrAuthorNotFound) Is(target error) bool {
    _, ok := target.(*NotFoundError)
    return ok
}

func main() {
    r := Repository{}
    author, err := r.FetchAuthor("isbn-213f-23422f52356")
    if err != nil {
        // Fetch the original and determine if the error is recoverable
        if error.Is(err, &ErrAuthorNotFound{}) {
            author, err := r.AddBook("isbn-213f-23422f52356", "charles", "darwin")
        }
        if err != nil {
            logrus.WithFields(errors.ToLogrus(err)).
				WithError(err).Error("while fetching author")
            os.Exit(1)
        }
    }
    fmt.Printf("Author %+v\n", author)
}
```