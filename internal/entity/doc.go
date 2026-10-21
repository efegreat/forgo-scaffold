// Package entity holds the business objects the domains share, such as orders and shipments. Their shape
// is the business's, not a table's: they carry no tags and no database types, and the adapters map them to
// whatever tables store them. A domain can read the entities of another without importing it.
package entity
