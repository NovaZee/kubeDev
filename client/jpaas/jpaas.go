package jpaas

// JPASSClient is a client for accessing JPASS.

type Client interface {
}

type client struct {
}

func New() Client {
	return &client{}
}
