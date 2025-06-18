package query

import (
    "github.com/hiicl/GPU-over-IP-AC922/pkg/gpu"
)

// GPUUUIDsByIDs 根据 GPU 的 deviceID 列表返回其 UUID 列表
func GPUUUIDsByIDs(deviceIDs []int) []string {
    uuidList := []string{}
    allGPUs := ListGPUs() // 已有函数，返回所有 GPUInfo（包含 UUID 和 index）

    indexMap := make(map[int]string)
    for _, g := range allGPUs {
        indexMap[int(g.Index)] = g.Uuid
    }

    for _, id := range deviceIDs {
        if uuid, ok := indexMap[id]; ok {
            uuidList = append(uuidList, uuid)
        }
    }

    return uuidList
}
