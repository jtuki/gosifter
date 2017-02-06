package api

const (
	MAX_JSON_FIELD_NUMBER = 4096
)

const (
	TAG_CONFIDENTIAL           = "confidential"
	TAG_CONFIDENTIAL_SEPARATOR = ","
)

// 资源的保密级别
const (
	CONFIDENTIAL_LEVEL0 = 0 // 公开
	CONFIDENTIAL_LEVEL1 = 1 // 保密
	CONFIDENTIAL_LEVEL2 = 2 // 高度保密
	CONFIDENTIAL_LEVEL3 = 3 // 绝密

	_c_level_max           = CONFIDENTIAL_LEVEL3
	CONFIDENTIAL_LEVEL_MAX = _c_level_max
)

// confidential level tag
const (
	CLEVEL_TAG_OMIT   = "-"
	CLEVEL_TAG_LEVEL0 = "level0"
	CLEVEL_TAG_LEVEL1 = "level1"
	CLEVEL_TAG_LEVEL2 = "level2"
	CLEVEL_TAG_LEVEL3 = "level3"
)

// @refer parseConfidentialTags()
var _confidential_map map[string]int = map[string]int{
	CLEVEL_TAG_OMIT:   CONFIDENTIAL_LEVEL0,
	CLEVEL_TAG_LEVEL0: CONFIDENTIAL_LEVEL0,
	CLEVEL_TAG_LEVEL1: CONFIDENTIAL_LEVEL1,
	CLEVEL_TAG_LEVEL2: CONFIDENTIAL_LEVEL2,
	CLEVEL_TAG_LEVEL3: CONFIDENTIAL_LEVEL3,
}
