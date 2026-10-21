// Package entity holds the business objects the domains share, such as orders and shipments. Their shape
// is the business's, not a table's: they carry no tags and no database types, and the adapters map them to
// whatever tables store them. A domain can read the entities of another without importing it.
//
// An entity's methods keep the rules that need nothing outside it, such as Order.Ship, which only ships a
// placed order. Rules that need a port or a live config belong to the services of the domains.
package entity
