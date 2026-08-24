package utils

func GenerateShortCode(id uint64) string {
	// Base62字符集：0-9, a-z, A-Z（标准顺序）
	const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const offset = 100000 // 偏移量，让短码从4位开始
	id = id + offset
	if id == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	base := uint64(len(base62Chars)) // base = 62

	// 第3步：核心循环: 短除法   所有进制转换的通用数学方法
	for id > 0 {
		remainder := id % base                          // 取余数
		result = append(result, base62Chars[remainder]) // 余数对应字符
		id = id / base                                  // 整除，进入下一位
	}

	// 第4步：反转结果   首尾交换位置
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func GenerateJWT(pwd string) string {
	return pwd
}
