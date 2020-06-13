package datanode

//DataNode Represents a data storage node in the system
type DataNode struct {
	IP           string       //IP of the data node host
	InternalPort string       //Port on which all internal comm is done
	NameNode     NameNodeData //Houses the needed info about the current name node
}

//NameNodeData Houses the needed info about the name node
type NameNodeData struct {
	IP   string //IP of the name node host
	Port string //Port on which the data node communicates with the name node
}
