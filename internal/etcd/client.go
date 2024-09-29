package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

// 客户端对象
type Client struct {
	Client3    *clientv3.Client
	kv         clientv3.KV
	lease      clientv3.Lease
	watch      clientv3.Watcher
	serverList map[string]string
	lock       sync.Mutex
}

var Clients *Client

// 初始化客户端对象
func InitClient(addr []string) error {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(conf)
	if err != nil {
		fmt.Printf("create connection etcd failed %s\n", err)
		return err
	}

	// 得到 KV 、Lease、 Watcher 的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watch := clientv3.NewWatcher(client)

	// 给客户端对象赋值

	Clients = &Client{
		Client3:    client,
		kv:         kv,
		lease:      lease,
		watch:      watch,
		serverList: make(map[string]string),
	}

	return nil
}

// 根据注册的服务名，获取服务实例的信息
func (c *Client) getServiceByName(prefix string) ([]string, error) {
	// 读取的时候带有 WithPrefix 选项，所以会读取该前缀所有的字段值
	resp, err := c.kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		fmt.Printf("getServiceByName failed %s\n", err)
		return nil, err
	}
	// 返回的 resp 是多个字段值。需要遍历提取对应的 key value
	addrs := c.extractAddrs(resp)
	return addrs, nil

}

// 根据 etcd 的响应，提取服务实例的数组
func (c *Client) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}

	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			// 将 key  value 值保存在  ServiceList 表中
			c.SetServiceList(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

// 设置 serverList
func (c *Client) SetServiceList(key, val string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// serverList 为初始化设置的本地 map 对象，由于考虑到多个 Client3 运行，所以需要加锁控制
	c.serverList[key] = string(val)
	fmt.Println("set data key :", key, "val:", val)
}

// 删除本地缓存的服务实例信息
func (c *Client) DelServiceList(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.serverList, key)
	fmt.Println("del data key:", key)

	newRes, err := c.getServiceByName(key)
	if err != nil {
		fmt.Printf("getServiceByName failed %s\n", err)
	} else {
		fmt.Printf("get  key %s", key, " current val is: %v\n", newRes)
	}

}

// 获取服务实例信息
func (c *Client) GetService(prefix string) ([]string, error) {
	if addrs, err := c.getServiceByName(prefix); err != nil {
		panic(err)
	} else {
		fmt.Println("get service ", prefix, " for instance list: ", addrs)
		go c.watcher(prefix)
		return addrs, nil
	}
}

// 监控指定键值对的变更
func (c *Client) watcher(prefix string) {
	watchRespChan := c.watch.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT: // 写入的事件
				c.SetServiceList(string(event.Kv.Key), string(event.Kv.Value))
			case mvccpb.DELETE: // 删除的事件
				c.DelServiceList(string(event.Kv.Key))
			}
		}
	}
}
