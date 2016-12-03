package coral

// 系统保留错误状态码
const STATUS_SUCCESS = 0        // 成功
const STATUS_ERROR_UNKNOWN = 1  // 未指定异常
const STATUS_ERROR_DB = 2       // 数据库异常
const STATUS_INVALID_PARAM = 3  // 参数校验异常
const STATUS_INVALID_STATUS = 4 // 输出status超出预期
