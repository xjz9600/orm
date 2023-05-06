package reflect

import (
	"reflect"
)

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)
	numMethod := typ.NumMethod()
	res := make(map[string]FuncInfo, numMethod)
	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func
		numIn := fn.Type().NumIn()
		input := make([]reflect.Type, 0, numIn)
		inputValues := make([]reflect.Value, 0, numIn)
		input = append(input, reflect.TypeOf(entity))
		inputValues = append(inputValues, reflect.ValueOf(entity))
		for i := 1; i < numIn; i++ {
			fnOnType := fn.Type().In(i)
			input = append(input, fnOnType)
			inputValues = append(inputValues, reflect.Zero(fnOnType))
		}
		numOut := fn.Type().NumOut()
		output := make([]reflect.Type, 0, numOut)
		for i := 0; i < numOut; i++ {
			output = append(output, fn.Type().Out(i))
		}
		resValues := fn.Call(inputValues)
		result := make([]any, 0, len(resValues))
		for _, v := range resValues {
			result = append(result, v.Interface())
		}
		res[method.Name] = FuncInfo{
			Name:       method.Name,
			InputType:  input,
			OutputType: output,
			Result:     result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name       string
	InputType  []reflect.Type
	OutputType []reflect.Type
	Result     []any
}
