package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

type AppNotFoundError struct {
	specific string
}

func (e AppNotFoundError) Error() string {
	return "Application not found for name=" + e.specific
}
