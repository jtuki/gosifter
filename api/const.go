package api

// 资源的保密级别
const (
	CONFIDENTIAL_LEVEL0 = 0 // 公开
	CONFIDENTIAL_LEVEL1 = 1 // 保密
	CONFIDENTIAL_LEVEL2 = 2 // 高度保密
	CONFIDENTIAL_LEVEL3 = 3 // 绝密

	_c_level_max           = CONFIDENTIAL_LEVEL3
	CONFIDENTIAL_LEVEL_MAX = _c_level_max
)
