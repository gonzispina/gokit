# gokit

My personal golang kit to use in different projects. 

### What you will find?
Most of the packages are wrappers to encapsulate libraries and configurations in one place. 
Others are to improve APIs or to add some missing functionality.

## Packages

### Context

Improves the functionality of the standard's lib context. Every context
is created with its own tracking id that can be used for logging. 

- Adds a method to the standard library context's interface: "TrackingID"
- Exports a function "WithID" to derive a context with its trackingID
- Exports a function "Upgrade" to create a new context from one of the standard's lib
- Exports a function "Merge" to merge a standard's lib context with an existing context of this package

### Logs

Is a wrapper around the zap logger created by Uber. It has some extra features:

- Exports "UserID" function to add a key the logger's methods with a user id
- Exports "Error" function to add a key the logger's methods with an error
- It expects a context, from this same module, in every method and adds the tracking id to every log produced.
- It initializes a zap logger with a default config and a "default" namespace that is added to every log produced.
- A logger with a new namespace can be derived from a previous logger

### UUID

Is a wrapper around google's uuid creator. It exports a single method "New" which
returns a string uuid. 
