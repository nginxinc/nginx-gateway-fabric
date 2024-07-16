# Go Style Guide

Before diving into project-specific guidelines, please familiarize yourself with the following vetted best practices:

- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [100 Go Mistakes](https://github.com/teivah/100-go-mistakes)

Once you have a good grasp of these general best practices, you can then explore the project-specific guidelines below.
These guidelines will often build upon the foundation set by the general best practices and provide additional
recommendations tailored to the project's specific requirements and coding style.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Table of Contents

- [General Guidelines](#general-guidelines)
  - [Use the empty struct `struct{}` for sentinel values](#use-the-empty-struct-struct-for-sentinel-values)
  - [Consistent Line Breaks](#consistent-line-breaks)
  - [Do not copy sync entities](#do-not-copy-sync-entities)
  - [Construct slices with known capacity](#construct-slices-with-known-capacity)
  - [Accept interfaces and return structs](#accept-interfaces-and-return-structs)
  - [Use contexts in a viral fashion](#use-contexts-in-a-viral-fashion)
  - [Do not use templates to replace interface types](#do-not-use-templates-to-replace-interface-types)
  - [Do not use booleans as function parameters](#do-not-use-booleans-as-function-parameters)
  - [Use dependency injection to separate concerns](#use-dependency-injection-to-separate-concerns)
  - [Required arguments should be provided via parameters and optional arguments provided functionally or with structs](#required-arguments-should-be-provided-via-parameters-and-optional-arguments-provided-functionally-or-with-structs)
- [Error Handling](#error-handling)
  - [Prefer inline error handling](#prefer-inline-error-handling)
  - [Do not filter context when returning errors](#do-not-filter-context-when-returning-errors)
  - [Only handle errors once](#only-handle-errors-once)
  - [Libraries should return errors for callers to handle](#libraries-should-return-errors-for-callers-to-handle)
  - [Callers should handle errors and pass them up the stack for notification](#callers-should-handle-errors-and-pass-them-up-the-stack-for-notification)
  - [Use panics for unrecoverable errors or programming errors](#use-panics-for-unrecoverable-errors-or-programming-errors)
- [Logging](#logging)
- [Concurrency](#concurrency)
- [Recommended / Situational](#recommended--situational)
  - [Use golang benchmark tests and pprof tools for profiling and identifying hot spots](#use-golang-benchmark-tests-and-pprof-tools-for-profiling-and-identifying-hot-spots)
  - [Reduce the number of stored pointers. Structures should store instances whenever possible](#reduce-the-number-of-stored-pointers-structures-should-store-instances-whenever-possible)
  - [Pass pointers down the stack not up](#pass-pointers-down-the-stack-not-up)
  - [Using interface types will cause unavoidable heap allocations](#using-interface-types-will-cause-unavoidable-heap-allocations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## General Guidelines

### Use the empty struct `struct{}` for sentinel values

Empty structs as sentinels unambiguously signal an explicit lack of information. For example, use empty struct for sets
and for signaling via channels that don't require a message.

DO:

```go
set := make(map[string]struct{}) // empty struct is nearly value-less

signaller := make(chan struct{}, 0)
signaller <- struct{}{} // no information but signal on delivery
```

DO NOT:

```go
set := make(map[string]bool) // is true/false meaningful? Is this a set?

signaller := make(chan bool, 0)
signaller <- false // is this a signal? is this an error?
```

### Consistent Line Breaks

When breaking up a (1) long function definition, (2) call, or (3) struct initialization, choose to break after each parameter,
argument, or field.

DO:

```go
// 1
func longFunctionDefinition(
  paramX int,
  paramY string,
  paramZ bool,
) (string, error){}

// 2
callWithManyArguments(
    arg1,
    arg2,
    arg3,
)

// 3
s := myStruct{
  field1: 1,
  field2: 2,
  field3: 3,
}
```

DO NOT:

```go
// 1
func longFunctionDefinition(paramX int, paramY string,
  paramZ bool,
) (string, error){}

// or

func longFunctionDefinition(
  paramX int, paramY string,
  paramZ bool,
) (string, error){}

// 2
callWithManyArguments(arg1, arg2,
    arg3)

// 3
s := myStruct{field1: 1, field2: 2,
field3: 3}

```

> **Exception**: Calls to ginkgo helper functions like below in tests are OK (because of the readability):
>
>```go
>DescribeTable("Edge cases for events",
>    func(e interface{}) {
>    . . .
>)
>```

When constructing structs pass members during initialization.

Example:

```go
cfg := foo.Config{
  Site: "example.com",
  Out: os.Stdout,
  Dest: c.KeyPair{
    Key: "style",
    Value: "well formatted",
  },
}
```

### Do not copy sync entities

`sync.Mutex` and `sync.Cond` MUST NOT be copied. By extension, structures holding an instance MUST NOT be copied, and
structures which embed instances MUST NOT be copied.

DO NOT embed sync entities. Pointers to `Mutex` and `Cond` are required for storage.

### Construct slices with known capacity

Whenever possible bounded slices should be constructed with a length of zero size, but known capacity.

```go
s := make([]string, 0, 32)
```

Growing a slice is an expensive deep copy operation. When the bounds of a slice can be calculated, pre-allocating the
storage allows for append to assign a value without allocating new memory.

### Accept interfaces and return structs

Structures are expected to be return values from functions. If they satisfy an interface, any struct may be used in
place of that interface. Interfaces will cause escape analysis and likely heap allocation when returned from a
function - concrete instances (such as copies) may stay as stack memory.

Returning interfaces from functions will hide the underlying structure type. This can lead to unintended growth of the
interface type when methods are needed but unavailable, or an API change to return the structure later.

Accepting interfaces as arguments ensures forward compatibility as API responsibilities grow. Structures as arguments
will require additional functions or breaking API changes. Whereas interfaces inject behavior and may be replaced or
modified without changing the signature.

### Use contexts in a viral fashion

Functions that accept a `context.Context` should pass the context or derive a subsidiary context to functions it calls.
When designing libraries, subordinate functions (especially those asynchronous or expensive in nature) should accept
`context.Context`.

Public APIs SHOULD be built from inception to be context aware. ALL asynchronous public APIs MUST be built from
inception to be context aware.

### Do not use templates to replace interface types

Use _templates_ to substitute for concrete types, use _interfaces_ to substitute for abstract behaviors.

### Do not use booleans as function parameters

A boolean can only express an on/off condition and will require a breaking change or new arguments in the future.
Instead, pack functionality; for example, using integers instead of bools (maps and structs are also acceptable).

DO NOT:

```go
func(bool userOn, bool groupOn, bool globalOn)
```

DO:

```go
func(uint permissions) // permissions := USER | GROUP | GLOBAL;  if permissions & USER then USER is set
```

### Use dependency injection to separate concerns

Use dependency injection to separate high-level policy from low-level detail. Dependency injection promotes modular,
testable, and maintainable code by reducing coupling and increasing flexibility.

Creating dependencies couples ownership and lifetime while making tests difficult or impossible. Constructing
dependencies adds side-effects which complicates testing.

### Required arguments should be provided via parameters and optional arguments provided functionally or with structs

DO NOT:

```go
func(int required, int optional) {
  if optional {...}
}
```

DO:

```go
type Option func (o *Object)

func Optional(string optional) Option {
  return func (o *Object) {
    o.optional = optional
  }
}

func (int required, ...Options) {
  for o := range Options {
    o(self)
  }
}
```

## Error Handling

### Prefer inline error handling

When possible, use inline error handling.

DO:

```go
if err := execute(); err != nil {
    // handle error
}
```

DO NOT:

```go
err := execute()
if err != nil {
    // handle error
}
```

### Do not filter context when returning errors

Preserve error context by wrapping errors as the stack unwinds. Utilize native error wrapping with `fmt.Errorf` and
the `%w` verb to wrap errors. Wrapped errors offer a transparent view to end users. For a practical example, refer to
this runnable code snippet: [Go Playground Example](https://go.dev/play/p/f9EaJDB5JUO). When required, you can identify
inner wrapped errors using native APIs such as `As`, `Is`, and `Unwrap`.
See [Working with errors](https://go.dev/blog/go1.13-errors) for more information on error wrapping.

### Only handle errors once

DO NOT log an error, then subsequently return that error. This creates the potential for multiple error reports and
conflicting information to users.

DO NOT:

```go
func badAtStuff(noData string) error {
  if len(noData) == 0 {
    fmt.Printf("Received no data")
  }

  return errors.New("received no data")
}
```

DO

```go
func badAtStuff(noData string) error {
  if len(noData) == 0 {
    return errors.New("received no data")
  }
  ...
}
```

A leader of the golang community once said:
> Lastly, I want to mention that you should only handle errors once. Handling an error means inspecting the
> error value, and making a decision. If you make less than one decision, you’re ignoring the error...But making
> more than one decision in response to a single error is also problematic. - Dave Cheney

### Libraries should return errors for callers to handle

Asynchronous libraries should communicate via channels or callbacks. Only then should they log unhandled errors.

Example:

```go
func onError(err error) {
// got an asynchronous error
}

func ReadAsync(r io.Reader, onError) {
  err := r()
  if err != nil {
    onError(err)
  }
}

go ReadAsync(reader, onError)

// OR errs := make(chan error)

func ReadAsync(r io.Reader, errs chan<- error) {
  err := r()
  if err != nil {
    // put error on errs channel, but don't block forever.
  }
}
```

### Callers should handle errors and pass them up the stack for notification

Callers should handle errors that occur within the functions they call. This allows them to handle errors according to
their specific requirements. However, if callers are unable to handle the error or need to provide additional context,
they can add context to the error and pass it up the stack for notification.

Example:

```go
func readFile(filename string) ([]byte, error) {
  file, err := os.Open(filename)
  if err != nil {
    return nil, fmt.Errorf("failed to open file: %w", err)
  }
  defer file.Close()

  data, err := ioutil.ReadAll(file)
  if err != nil {
    return nil, fmt.Errorf("failed to read file: %w", err)
  }

  return data, nil
}

func processFile(filename string) error {
  data, err := readFile(filename)
  if err != nil {
    return fmt.Errorf("failed to process file: %w", err)
  }
  // Process the file data here
  return nil
}

func main() {
  filename := "example.txt"
  err := processFile(filename)
  if err != nil {
    fmt.Printf("Error processing file: %v\n", err) // caller handles the error
  }
}
```

### Use panics for unrecoverable errors or programming errors

Panics should be used in the following cases:

1. Unrecoverable errors. An unrecoverable error is when NGF cannot continue running or its behavior or internal state
   cannot be guaranteed. One example of this is if an error occurs when adding the Kubernetes API types to the Scheme,
   or if an error occurs when marking a CLI flag as required.
2. Programming errors. A programming error is an error that is only possible if there was a programming mistake. For
   example, if the wrong type is passed or a go template is passed an invalid value.

When using panics, pass an error as the argument. For example:

```go
panic(fmt.Errorf("unknown event type %T", e))
```

## Logging

See the [Logging guidelines](logging-guidelines.md) document.

## Concurrency

Below are some general guidelines to follow for writing concurrent code:

- **Don't assume that a concurrent solution will be faster than an iterative one**: Benchmark the iterative and
  concurrent solutions before deciding which one to use.
- **Don't add synchronization unless strictly necessary**: Synchronization primitives -- such as mutexes -- are costly.
  Only use them when the code is accessed by multiple goroutines concurrently.
- **Document when exported code is not concurrent-safe**: If an exported interface, object, or function is not
  reentrant, you _must_ document that in the comments. Make it clear and obvious.
- **Don't leak goroutines**: Goroutines are not garbage collected by the runtime, so every goroutine you start must also
  be cleaned up. Here's a couple of related principles:
  - "If a goroutine is responsible for creating a goroutine, it is also responsible for ensuring it can stop the
      goroutine." -- [Concurrency in Go][cig]
  - "Before you start a goroutine, always know when, and how, it will stop." -- [Concurrency Made Easy][cheney].
- **Blocking operations within a goroutine must be preemptable**: This allows goroutines to be cancelled and prevents
  goroutine leaks.
- **Leverage contexts**: Contexts allow you to enforce deadlines and send cancellation signals to multiple goroutines.
- **Avoid buffered channels**:  Use unbuffered channels unless there is a very good reason for using a buffered channel.
  Unbuffered channels provide strong synchronization guarantees. Buffered channels are asynchronous and will not block
  unless the channel is full. Buffered channels can also be slower than unbuffered channels.
- **Protect maps and slices**: Maps and slices cannot be accessed concurrently (when at least one goroutine is writing)
  without locking. Doing so can lead to data races.
- **Never copy sync types**: see [above section](#do-not-copy-sync-entities).
- **Choose primitives or channels based on use case**: In general, the Go language writers tell us to prefer channels
  and communication for synchronization over primitives in the sync package such as mutexes and wait groups. "Do not
  communicate by sharing memory. Instead, share memory by communicating". However, in practice, there are some cases
  where sync primitives are the better choice. For example, you should use primitives if you are working with a
  performance-critical section, or protecting the internal state of a struct. However, you should use channels if you
  are transferring ownership of data (e.g. producer/consumer) or trying to coordinate multiple pieces of logic.
- **When possible, write code that is implicitly concurrent-safe**: Code that is implicitly concurrent-safe can be
  safely accessed by multiple goroutines concurrently without any synchronization. For example, immutable data is
  implicitly concurrent-safe. Concurrent processes can operate on the data, but they can't modify it. Another example is
  lexical confinement. From [Concurrency in Go][cig]: "Lexical confinement involves using lexical scope to expose only
  the correct data and concurrency primitives for multiple concurrent processes to use. It makes it impossible to do the
  wrong thing." Code that is implicitly concurrent-safe is typically more performant and easier for developers to
  understand.
- **Leverage errgroup**: The package [errgroup]((https://pkg.go.dev/golang.org/x/sync/errgroup)) helps synchronize a
  group of goroutines that return errors.
- **Release locks and semaphores in the reverse order you acquire them**: This will prevent lock inversion and
  deadlocks.
- **Close channels to signal receivers NOT to free resources**: Channels do not need to be closed to free resources.
  Only close channels as a means to signal the channel's receivers that the channel is done accepting new data.

Sources and Recommended Reading:

- [Concurrency in Go][cig]
- [Concurrency Made Easy][cheney]
- [Channel Axioms](https://dave.cheney.net/2014/03/19/channel-axioms)
- [Concurrency is not parallelism](https://go.dev/blog/waza-talk)
- [Pipelines and cancellation](https://go.dev/blog/pipelines)
- [100 Go Mistakes; Chapters 8 & 9](https://github.com/teivah/100-go-mistakes)

[cig]:https://learning.oreilly.com/library/view/concurrency-in-go/9781491941294/

[cheney]: https://dave.cheney.net/paste/concurrency-made-easy.pdf

## Recommended / Situational

These recommendations are generally related to performance and efficiency but will not be appropriate for all paradigms.

### Use golang benchmark tests and pprof tools for profiling and identifying hot spots

The `-gcflags '-m'` can be used to analyze escape analysis and estimate logging costs.

### Reduce the number of stored pointers. Structures should store instances whenever possible

DO NOT use pointers to avoid copying. Pass by value. Ancillary benefit is reduction of nil checks. Fewer pointers helps
garbage collection and can indicate memory regions that can be skipped. It reduces de-referencing and bounds checking in
the VM. Keep as much on the stack as possible (see caveat: [accept interfaces](#accept-interfaces-and-return-structs) -
forward compatibility and flexibility concerns outweigh costs of heap allocation)

FAVOR:

```go
type Object struct{
  subobject SubObject
}

func New() Object {
  return Object{
    subobject: SubObject{},
  }
}
```

DISFAVOR:

```go
type Object struct{
  subobject *SubObject
}

func New() *Object {
  return &Object{
    subobject: &SubObject{},
  }
}
```

### Pass pointers down the stack not up

Pointers can be passed down the stack without causing heap allocations. Passing pointers up the stack will cause heap
allocations.

```text
-> initialize struct A -> func_1(&A)
-> func_2(&A)
```

`A` can be passed as a pointer as it's passed down the stack.

Returning pointers can cause heap allocations and should be avoided.

DO NOT:

```go
func(s string) *string {
  s := s + "more strings"
  return &s // this will move to heap
}
```

### Using interface types will cause unavoidable heap allocations

Frequently created, short-lived instances will cause heap and garbage collection pressure. Using a `sync.Pool` to store
and retrieve structures can improve performance.

Before adopting a `sync.Pool` analyze the frequency of creation and duration of lifetime. Objects created frequency for
short periods of time will benefit the most.

For example, channels that send and receive signals using interfaces; these are often only referenced for the duration
of the event and can be recycled once the signal is received and processed.
