package health

// Storage 定义健康状态存储接口
//
// 该接口抽象了健康状态的存储操作，允许外部实现不同的存储策略：
//   - 内存存储（如 sync.Map、普通 map + 锁）
//   - 数据库存储（如 MySQL、PostgreSQL）
//   - 缓存存储（如 Redis、Memcached）
//   - 混合存储（如缓存 + 数据库）
//
// 实现者需要保证线程安全性
type Storage interface {
	// Get 获取指定资源的健康状态
	//
	// 参数：
	//   - resourceType: 资源类型
	//   - resourceID: 资源 ID
	//
	// 返回值：
	//   - *Health: 健康状态对象，如果不存在返回 nil
	//   - error: 错误信息
	Get(resourceType ResourceType, resourceID uint) (*Health, error)

	// Set 设置指定资源的健康状态
	//
	// 如果资源已存在，则更新；如果不存在，则创建
	//
	// 参数：
	//   - status: 健康状态对象
	//
	// 返回值：
	//   - error: 错误信息
	Set(status *Health) error

	// Delete 删除指定资源的健康状态
	//
	// 参数：
	//   - resourceType: 资源类型
	//   - resourceID: 资源 ID
	//
	// 返回值：
	//   - error: 错误信息
	Delete(resourceType ResourceType, resourceID uint) error
}
