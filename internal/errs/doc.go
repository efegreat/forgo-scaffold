// Package errs holds the business errors of every domain, as operation.Error values. Their codes are the
// error codes of the API, which clients match on, and errors.Is matches them by code, so a code must name
// one error in the whole application: one package keeps them unique, and in one place to document.
package errs
