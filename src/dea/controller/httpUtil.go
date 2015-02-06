package controller

import (
	"net/http"
	"encoding/json"
)

//  响应http请求
func writeJson(data []byte, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type","application/json")
	rw.Write(data)
}

func writeStr(data string, rw http.ResponseWriter ) {
	rw.Header().Set("Content-Type","application/x-drw")
	rw.Write([]byte(data))
}


//将对象转换成json格式
func encodeJson(v interface{}) ([] byte, error){
	
	r,err := json.Marshal(v)
	
	if err != nil {
		return nil,err
	}
	
	return r,nil
}