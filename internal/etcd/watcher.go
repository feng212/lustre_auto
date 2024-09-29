package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"lustre_auto/config"
	"lustre_auto/internal/resource"
)

// Node 控制器节点结构体，描述控制器的名称、IP、优先级和状态
type Node struct {
	Name     string `json:"name"`     // 节点名称
	IP       string `json:"ip"`       // 节点 IP 地址
	Priority int    `json:"priority"` // 节点优先级
	Status   string `json:"status"`   // 节点状态：Healthy, Unreachable, Failed , running
}

// Resource 资源结构体，描述资源的挂载状态及优先级
type Resource struct {
	Name        string `json:"name"` // 资源名称
	A           Node   `json:"A"`    // 主节点（如：controllerA）
	B           Node   `json:"B"`    // 备份节点（如：controllerB）
	ResName     string `json:"res_name"`
	MountPoint  string `json:"mount_point"`  // 资源挂载点
	CurrentNode string `json:"current_node"` // 当前挂载的节点
	Status      string `json:"status"`       // 资源状态：mounted/unmounted
}

// watchResources 监听与本控制器相关的资源状态变化
func (c *Client) watchResources(controllerName string) {
	prefix := fmt.Sprintf("/resources/%s", controllerName)
	watchChan := c.watch.Watch(context.Background(), prefix, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Printf("Resource updated: %s %s\n", event.Kv.Key, event.Kv.Value)
				c.handleResourceChange(event.Kv.Key, event.Kv.Value)
			case clientv3.EventTypeDelete:
				fmt.Printf("Resource deleted: %s\n", event.Kv.Key)
			}
		}
	}
}

// handleResourceChange 处理资源状态变化
func (c *Client) handleResourceChange(key, value []byte) {
	fmt.Printf("Handling resource change for key: %s, new value: %s\n", key, value)

	var res *Resource
	err := json.Unmarshal(value, &res)
	if err != nil {
		log.Printf("Error unmarshalling resource: %s\n", err)
		return
	}

	// 根据优先级和节点状态决定挂载在哪个节点上
	c.processResource(res)
}

// processResource 根据优先级和节点状态决定资源的挂载位置
func (c *Client) processResource(res *Resource) {
	primaryNode, backupNode := &Node{}, &Node{}
	if res.A.Priority > res.B.Priority {
		primaryNode, backupNode = &res.A, &res.B
	} else {
		primaryNode, backupNode = &res.B, &res.A
	}

	if primaryNode.Status != "Healthy" && backupNode.Status != "Healthy" {
		return
	}

	// 尝试切换到优先级较高的节点
	if c.NodeHealth(res.Name, primaryNode.IP) {
		if res.Status == "mounted" && primaryNode.Status == "running" {
			return
		}
		if res.CurrentNode != primaryNode.Name && primaryNode.IP == config.ConfigData.Server.IP {
			fmt.Printf("Switching resource %s to %s\n", res.Name, primaryNode.Name)
			err := c.switchResourceToNode(res, primaryNode.Name)
			if err != nil {
				//资源启动失败，修改status
				primaryNode.Status = "Failed"
				log.Printf("Resource %s failed to start on %s: %v\n", res.Name, primaryNode.Name, err)
			}
			primaryNode.Status = "running"

		}
	} else {
		primaryNode.Status = "Unreachable"
	}
	if c.NodeHealth(res.Name, backupNode.IP) { //判断首选节点状态，若不健康则切换到备份节点
		// 如果优先节点不健康，切换到备份节点
		if res.Status == "mounted" && backupNode.Status == "running" {
			return
		}
		if res.CurrentNode != backupNode.Name && backupNode.IP == config.ConfigData.Server.IP {
			fmt.Printf("Switching resource %s to %s\n", res.Name, backupNode.Name)
			err := c.switchResourceToNode(res, backupNode.Name)
			if err != nil {
				backupNode.Status = "Failed"
				log.Printf("Resource %s failed to start on %s: %v\n", res.Name, primaryNode.Name, err)
			}
			backupNode.Status = "running"
		}
	} else {
		backupNode.Status = "Unreachable"
	}

	err := c.updateResourceStatus(res)
	log.Printf("Failed to update resource status after 3 attempts: %s\n", err)
}

// NodeHealth 检查指定节点的健康状态
func (c *Client) NodeHealth(resourceName, nodeIP string) bool {
	key := fmt.Sprintf("/controller/%s/%s", resourceName, nodeIP)
	_, err := c.getServiceByName(key)
	if err != nil {

		log.Printf("Error checking node health: %s\n", err)
		return false
	}
	return true
}

// switchResourceToNode 将资源切换到指定的节点
func (c *Client) switchResourceToNode(res *Resource, targetNode string) error {
	fmt.Printf("Switching resource %s to node: %s\n", res.Name, targetNode)

	// 更新资源状态为挂载在目标节点
	// 最多重试3次
	var err error
	for i := 0; i < 3; i++ {
		err = resource.Mount(res.ResName, res.MountPoint, res.Name)
		if err == nil {
			res.CurrentNode = targetNode
			res.Status = "mounted"
			return nil
		}
		log.Printf("Failed to update resource status, retrying... (%d/3)\n", i+1)
	}
	res.CurrentNode = ""
	res.Status = "umounted"
	return err
}

// updateResourceStatus 更新资源状态到 etcd
func (c *Client) updateResourceStatus(res *Resource) error {
	key := fmt.Sprintf("/resources/%s", res.Name)
	value, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}
	var putErr error
	putErr = Service.PutService(key, string(value))
	if putErr == nil {
		fmt.Printf("Successfully updated resource status in etcd: %s\n", res.Name)
		return nil
	}
	return fmt.Errorf("%w", putErr)
}
