// Package postgres implements the storage ports on PostgreSQL, one file per port. Every statement runs
// through sqltx.From, so it joins the transaction its context carries: that of a service's
// operation.Transactor, or the consumer's, which commits a handler's writes together with the inbox.
//
// Each constructor returns its port, not its type. That checks at compile time that the type implements
// the port, binds it under the port, and lets a test put a fake in its place with forgotest.Replace.
package postgres
