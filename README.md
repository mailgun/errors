# Errors
A convenience package to add additional context fields to errors without the need to
implement a custom error type each time you return an error. 

`errors.Wrap()` includes a stack trace so logging can report the exact location where
the error occurred.
`errors.WithStack()` for when you don't need a message, just a stack trace to where
the error occured.
`errors.WithFields{}.Wrap()` to attach additional fields to the error so when we log
the error up the call stack, we have as much information as possible.

Provides pass through access to the standard `errors.Is()`, `errors.As()`,
`errors.Unwrap()` so you don't need to import this package and the standard 
error package.

Convenience functions to extract all stack and field information from the error.
`errors.ToLogrus()` and `errors.ToMap()`
 
## Adding structured fields to an error
Wraps the original error while providing structured field data
```go
_, err := ioutil.ReadFile(fileName)
if err != nil {
        return errors.WithFields{"file": fileName}.Wrap(err, "read failed")
}
```

## Retrieving the structured fields
Using `errors.WithFields{}` stores the provided fields for later retrieval by upstream code or structured logging
systems
```go
// Pass to logrus as structured logging
logrus.WithFields(errors.ToLogrus(err)).Error("open file error")
```
Stack information on the source of the error is also included
```go
context := errors.ToMap(err)
context == map[string]interface{}{
      "file": "my-file.txt",
      "excValue": "open file error",
	  "excType": "*my_package.TestError"
      "excFuncName": "my_package.TestFunction"
      "excFileName": "/path/to/example.go"
}
```

## Can be used with standard errors.Unwrap() and errors.Is() and errors.As()
Errors wrapped with `errors.WithFields{}` are compatible with standard library introspection functions
```go
var ErrQuery := errors.New("query error")
wrap := errors.WithFields{"key1": "value1"}.Wrap(err, "message")
errors.Is(wrap, ErrQuery) // == true
```

## Proper Usage
The fields wrapped by `errors.WithFields{}` is not intended to be used to by code to decide how an error should be 
handled. It is a convenience where the failure is well known, but the context is dynamic. In other words, you know the
database returned an unrecoverable query error, but creating a new error type with the details of each query
error is overkill **ErrorFetchPage{}, ErrorFetchAll{}, ErrorFetchAuthor{}, etc...**

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
Now you can easily search your structured logs for errors related to a customer.

You should continue to create and inspect error types
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
        // Fetch the original Cause() and determine if the error is recoverable
        if error.Is(err, &ErrAuthorNotFound{}) {
                author, err := r.AddBook("isbn-213f-23422f52356", "charles", "darwin")
        }
        if err != nil {
                logrus.WithFields(errors.ToLogrus(err)).Errorf("while fetching author - %s", err)
                os.Exit(1)
        }
    }
    fmt.Printf("Author %+v\n", author)
}
```

## Fields for concrete error types
If the error implements the `errors.HasFields` interface the context can be retrieved
```go
fields, ok := err.(errors.HasFields)
if ok {
    fmt.Println(fields.Fields())
}
```

This makes it easy for error types to provide their context information.
 ```go
type ErrBookNotFound struct {
    ISBN string
}
// Implements the `HasFields` interface
func (e *ErrBookNotFound) func Fields() map[string]interface{} {
    return map[string]interface{}{
        "isbn": e.ISBN,
    }
 }
```
Now we can create the error and logrus knows how to retrieve the context
 
```go
func (* Repository) FetchBook(isbn string) (*Book, error) {
    var book Book
    err := r.db.Query("SELECT * FROM books WHERE isbn = ?").One(&book)
    if err != nil {
        return nil, ErrBookNotFound{ISBN: isbn}
    }
}

func main() {
    r := Repository{}
    book, err := r.FetchBook("isbn-213f-23422f52356")
    if err != nil {
        logrus.WithFields(errors.ToLogrus(err)).Errorf("while fetching book - %s", err)
        os.Exit(1)
    }
    fmt.Printf("Book %+v\n", book)
}
```


## A Complete example
The following is a complete example using
http://github.com/mailgun/logrus-hooks/kafkahook to marshal the context into ES
fields.

```go
package main

import (
    "log"
    "io/ioutil"

    "github.com/mailgun/holster/v4/errors"
    "github.com/mailgun/logrus-hooks/kafkahook"
    "github.com/sirupsen/logrus"
)

func OpenWithError(fileName string) error {
    _, err := ioutil.ReadFile(fileName)
    if err != nil {
            // pass the filename up via the error context
            return errors.WithFields{
                "file": fileName,
            }.Wrap(err, "read failed")
    }
    return nil
}

func main() {
    // Init the kafka hook logger
    hook, err := kafkahook.New(kafkahook.Config{
        Endpoints: []string{"kafka-n01", "kafka-n02"},
        Topic:     "udplog",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Add the hook to logrus
    logrus.AddHook(hook)

    // Create an error and log it
    if err := OpenWithError("/tmp/non-existant.file"); err != nil {
        // This log line will show up in ES with the additional fields
        //
        // excText: "read failed"
        // excValue: "read failed: open /tmp/non-existant.file: no such file or directory"
        // excType: "*errors.WithFields"
        // filename: "/src/to/main.go"
        // funcName: "main()"
        // lineno: 25
        // context.file: "/tmp/non-existant.file"
        // context.domain.id: "some-id"
        // context.foo: "bar"
        logrus.WithFields(logrus.Fields{
            "domain.id": "some-id",
            "foo": "bar",
            "err": err,
        }).Error("log messge")
    }
}
```
