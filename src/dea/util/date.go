package util

import (
	"time"
)

//格式化当前时间 yyyy-mm-dd hh:mm:ss
func NowTime () string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func FormateTime(time time.Time) string{
	return time.Format("2006-01-02 15:04:05")
}
