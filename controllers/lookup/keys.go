package lookup

type LookupKey struct {
	Tier string
	Id   string
}

type ClusterKey = struct {
	Namespace string
	Cluster   string
}
