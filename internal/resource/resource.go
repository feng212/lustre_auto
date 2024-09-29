package resource

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"lustre_auto/config"
	"lustre_auto/internal/etcd"
	"os/exec"
)

func Mount(device, mountPoint, prefix string) error {
	session, err := concurrency.NewSession(etcd.Clients.Client3, concurrency.WithTTL(config.ConfigData.Etcd.Leasettl))
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()
	// 创建分布式锁 会话自动续约
	lockKey := fmt.Sprintf("%s/%s/mount_initing", prefix, device)
	mutex := concurrency.NewMutex(session, lockKey)
	// 加锁（阻塞直到获取到锁）
	fmt.Println("Trying to acquire lock...")
	if err := mutex.Lock(context.TODO()); err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Println("Lock acquired!")

	defer func() {
		// 释放锁
		if err := mutex.Unlock(context.TODO()); err != nil {
			log.Fatalf("Failed to release lock: %v", err)
		}
		fmt.Println("Lock released.")
	}()

	cmd := exec.Command("mount", device, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to mount %s at %s: %s", device, mountPoint, err)
		return err
	}
	fmt.Printf("Successfully mounted %s at %s\nOutput: %s\n", device, mountPoint, output)
	return nil
}
