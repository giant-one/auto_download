package main

import (
	"flag"
	"fmt"
)
import "../unit"

func main()  {
	// 定义几个变量，用于接收命令行的参数值
	var user        string
	// &user 就是接收命令行中输入 -u 后面的参数值，其他同理
	flag.StringVar(&user, "u", "root", "账号，默认为root")
	// 解析命令行参数写入注册的flag里
	flag.Parse()

	// 加密密钥
	passphrase := "111111111111111111111111"
	password := unit.Encrypt(user, passphrase)
	// 输出结果
	fmt.Printf("输入：%v\n秘钥：%v\n",
		user, password)
}