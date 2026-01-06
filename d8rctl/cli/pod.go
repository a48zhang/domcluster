package cli

import (
	"encoding/json"
	"fmt"

	"d8rctl/daemon"
)

// PodList 列出所有连接的 domclusterd 节点
func PodList() error {
	if !daemon.IsRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	nodes, err := daemon.GetNodeList()
	if err != nil {
		return fmt.Errorf("failed to get node list: %w", err)
	}

	if len(nodes) == 0 {
		fmt.Println("No nodes connected")
		return nil
	}

	fmt.Printf("Connected nodes: %d\n\n", len(nodes))

	for nodeID, info := range nodes {
		infoMap, ok := info.(map[string]interface{})
		if !ok {
			fmt.Printf("Node ID: %s (invalid data format)\n\n", nodeID)
			continue
		}

		name, _ := infoMap["name"].(string)
		role, _ := infoMap["role"].(string)
		version, _ := infoMap["version"].(string)

		fmt.Printf("Node ID: %s\n", nodeID)
		fmt.Printf("  Name:    %s\n", name)
		fmt.Printf("  Role:    %s\n", role)
		fmt.Printf("  Version: %s\n\n", version)
	}

	return nil
}

// PodListJSON 以 JSON 格式列出所有连接的 domclusterd 节点
func PodListJSON() error {
	if !daemon.IsRunning() {
		fmt.Println(`{"error":"Daemon is not running"}`)
		return nil
	}

	nodes, err := daemon.GetNodeList()
	if err != nil {
		return fmt.Errorf("failed to get node list: %w", err)
	}

	data, _ := json.MarshalIndent(nodes, "", "  ")
	fmt.Println(string(data))

	return nil
}