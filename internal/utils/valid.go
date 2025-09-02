package utils

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

func GetValidMsg(err error, obj any) string {
	//	使用的时候需要传obj的指针
	getobj := reflect.TypeOf(obj)
	//	将obj接口断言为具体类型
	if errs, ok := err.(validator.ValidationErrors); ok {
		//	断言成功
		for _, e := range errs {
			//循环每一个错误信息
			//根据具体报错字段名，获取结构体的具体字段
			if f, exits := getobj.Elem().FieldByName(e.Field()); exits {
				msg := f.Tag.Get("msg")
				return msg
			}
		}
	}
	return err.Error()
}
