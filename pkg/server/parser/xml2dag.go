package parser

import (
	"errors"

	"github.com/beevik/etree"
	"k8s.io/klog/v2"
)

var (
	dag NodeSet
	doc *etree.Document
)

// build vertices of DAG
func buildNode(tasks []*etree.Element) {
	for _, node := range tasks {
		switch node.SelectAttrValue("custom", "none") {
		case DockerType:
			newNode := NewDockerNode()
			if err := newNode.Fillin(node); err != nil {
				klog.Errorln("Failed to fill node info")
				break
			}
			dag[newNode.GetName()] = newNode
		default:
			break
		}
	}
}

// Build the edge of the DAG
func buildFlow(flows []*etree.Element) {
	for _, val := range flows {
		sourceRef := val.SelectAttrValue("sourceRef", "none")
		targetRef := val.SelectAttrValue("targetRef", "none")
		// As long as one of the two edges doesn't exist, ignore the edge
		if sourceRef == "none" || targetRef == "none" {
			continue
		}
		dag[targetRef].AppendDep(sourceRef)
	}
}

// The entry function of the entire DAG building module,
// completes the initialization of the map,
// loading the etree package, building the DAG, and returning the workflow id
func Xml2Dag(xmlstr string) (string, *NodeSet, error) {
	dag = make(map[string]Node)
	doc = etree.NewDocument()
	if err := doc.ReadFromString(xmlstr); err != nil {
		klog.Errorln(err, "Read xmlstr error")
		return "", nil, err
	}
	process := doc.SelectElement("definitions").SelectElement("process")
	process_id := process.SelectAttrValue("id", "none")
	buildNode(process.SelectElements("task"))
	buildFlow(process.SelectElements("sequenceFlow"))
	// After DAG is built, map stores the set of all nodes
	if process_id == "none" {
		return "", nil, errors.New("Workflow id is not present")
	}
	return process_id, &dag, nil
}
