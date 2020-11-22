package links

// create link map type
type LinkerID string
type linkMap map[LinkerID]NodeLinker

var Links = make(map[LinkerID]NodeLinker)

// create NodeLinker interface
type NodeLinker interface {
	GetId() LinkerID
	AddNode() error
	RemoveNode()
}

// fn to add all links from NodeLinkers in link map
func AddLinks(links linkMap) {
	for _, link := range links {
		link.AddNode()
	}
}

func AddLink(link NodeLinker) {
	Links[link.GetId()] = link
}

func RemoveLink(link NodeLinker) {
	delete(Links, link.GetId())
}
