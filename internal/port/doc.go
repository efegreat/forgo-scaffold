// Package port declares what the domains need from outside, such as storage, as interfaces in the terms
// of the domains (entity, event). The domains depend on the ports, and the packages under adapter implement
// them, so a domain never knows which database or service is behind one. Files group the ports by kind:
// repository.go holds every repository.
package port
