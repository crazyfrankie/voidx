# ErrorX

A Go error handling library that provides structured error management with status codes, stack traces, and configurable error registration.

## Features

- **Status Code Support**: Associate errors with specific status codes
- **Stack Traces**: Automatic stack trace generation for better debugging
- **Error Registration**: Pre-register error codes with messages and configurations
- **Flexible Options**: Support for key-value parameters and extra metadata
- **Stability Tracking**: Mark errors that affect system stability
- **Error Wrapping**: Wrap existing errors with additional context

## Installation

```go
import "github.com/crazyfrankie/frx/errorx"
```

## Quick Start

### 1. Register Error Codes
You can choose written by yourself or use our gen tools: [gen](https://github.com/crazyfrankie/frx/tree/master/errorx/gen)

Create error code definitions in your project's `types/errno` package:

```go
package errno

import "github.com/crazyfrankie/frx/errorx/code"

const (
    ErrPermissionDenied = int32(1000001)
    ErrResourceNotFound = int32(1000002)
    ErrInvalidParameter = int32(1000003)
)

func init() {
    // Register error codes with messages and options
    code.Register(
        ErrPermissionDenied,
        "unauthorized access: {user}",
        code.WithAffectStability(false),
    )
    
    code.Register(
        ErrResourceNotFound,
        "resource not found: {resource}",
        code.WithAffectStability(true),
    )
    
    code.Register(
        ErrInvalidParameter,
        "invalid parameter: {param}",
        code.WithAffectStability(false),
    )
}
```

### 2. Create and Use Errors

```go
package main

import (
    "fmt"
	
    "github.com/crazyfrankie/frx/errorx"
	
    "your-project/types/errno"
)

func main() {
    // Create a new error with parameters
    err := errorx.New(errno.ErrPermissionDenied, 
        errorx.KV("user", "john_doe"),
        errorx.Extra("request_id", "req-123"),
    )
    
    // Wrap an existing error
    originalErr := fmt.Errorf("database connection failed")
    wrappedErr := errorx.WrapByCode(originalErr, errno.ErrResourceNotFound,
        errorx.KV("resource", "user_table"),
    )
    
    // Extract error information
    if statusErr, ok := err.(errorx.StatusError); ok {
        fmt.Printf("Code: %d\n", statusErr.Code())
        fmt.Printf("Message: %s\n", statusErr.Msg())
        fmt.Printf("Affects Stability: %v\n", statusErr.IsAffectStability())
        fmt.Printf("Extra Data: %v\n", statusErr.Extra())
    }
}
```

## API Reference

### Core Functions

#### `New(code int32, options ...Option) error`
Creates a new error with the specified status code and options.

#### `WrapByCode(err error, statusCode int32, options ...Option) error`
Wraps an existing error with a status code and additional context.

#### `Wrapf(err error, format string, args ...any) error`
Wraps an error with a formatted message.

### Options

#### `KV(k, v string) Option`
Adds a key-value parameter that can be used in error message templates.

#### `KVf(k, v string, a ...any) Option`
Adds a formatted key-value parameter.

#### `Extra(k, v string) Option`
Adds extra metadata to the error.

### Registration Functions

#### `code.Register(code int32, msg string, opts ...RegisterOptionFn)`
Registers an error code with its message template and options.

#### `code.WithAffectStability(affectStability bool) RegisterOptionFn`
Configures whether the error affects system stability metrics.

#### `code.SetDefaultErrorCode(code int32)`
Sets a default error code for the system.

### StatusError Interface

```go
type StatusError interface {
    error
    Code() int32
    Msg() string
    IsAffectStability() bool
    Extra() map[string]string
}
```

## Best Practices

1. **Organize Error Codes**: Group related error codes in separate packages (e.g., `types/errno/auth`, `types/errno/database`)

2. **Use Meaningful Codes**: Choose error codes that are unique and meaningful within your system

3. **Template Messages**: Use parameter placeholders in error messages for dynamic content

4. **Stability Flags**: Mark errors that indicate system issues with `WithAffectStability(true)`

5. **Stack Traces**: Use `ErrorWithoutStack()` when you need clean error messages for user-facing responses