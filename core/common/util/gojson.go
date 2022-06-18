package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Js struct {
	data interface{}
}

//	=====================================================================
//	function name: Json
//	function type: public
//	function receiver: na
//	Initialize the json configruation
//	=====================================================================
func Json(data []byte) *Js {
	j := new(Js)
	var f interface{}

	err := json.Unmarshal(data, &f)
	if err != nil {
		return j
	}
	j.data = f

	return j
}

//	=====================================================================
//	function name: Get
//	function type: public
//	function receiver: *Js
//	According to the key of the returned data information,return js.data
//	=====================================================================
func (j *Js) Get(key string) *Js {
	jo := new(Js)
	m := j.Getdata()
	if v, ok := m[key]; ok {
		jo.data = v
		return jo
	}
	jo.data = nil
	return jo
}

//	=====================================================================
//	function name: IsValid
//	function type: public
//	function receiver: *Js
//	validate the JSON data, return bool
//	=====================================================================
func (j *Js) IsValid() bool {
	if nil == j.data {
		return false
	} else {
		return true
	}
}

//	=====================================================================
//	function name: Getdata
//	function type: public
//	function receiver: *Js
//	return JSON data
//	=====================================================================
func (j *Js) Getdata() map[string]interface{} {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m
	}

	return nil
}

//	=====================================================================
//	function name: Getindex
//	function type: public
//	function receiver: *Js
//	get JSON data index
//	=====================================================================
func (j *Js) Getindex(i int) *Js {
	jo := new(Js)
	num := i - 1
	if m, ok := (j.data).([]interface{}); ok {
		if num < len((j.data).([]interface{})) {
			v := m[num]
			jo.data = v
		} else {
			jo.data = nil
		}

		return jo
	}

	if m, ok := (j.data).(map[string]interface{}); ok {
		var n = 0
		var data = make(map[string]interface{})
		for i, v := range m {
			if n == num {
				switch vv := v.(type) {
				case float64:
					data[i] = strconv.FormatFloat(vv, 'f', -1, 64)
					jo.data = data
					return jo
				case string:
					data[i] = vv
					jo.data = data
					return jo
				case []interface{}:
					jo.data = vv
					return jo
				}
			}
			n++
		}
	}
	jo.data = nil

	return jo
}

//	=====================================================================
//	function name: Arrayindex
//	function type: public
//	function receiver: *Js
//	When the data {"result":["username","password"]} can use arrayindex(1)
//	get the username
//	=====================================================================
func (j *Js) Arrayindex(i int) string {
	num := i - 1
	if i > len((j.data).([]interface{})) {
		data := errors.New("index out of range list").Error()

		return data
	}

	if m, ok := (j.data).([]interface{}); ok {
		v := m[num]
		switch vv := v.(type) {
		case float64:
			return strconv.FormatFloat(vv, 'f', -1, 64)
		case string:
			return vv
		default:
			return ""
		}

	}

	if _, ok := (j.data).(map[string]interface{}); ok {
		return "error"
	}

	return "error"
}

//	=====================================================================
//	function name: Getkey
//	function type: public
//	function receiver: *Js
//	The data must be []interface{} ,According to your custom number to
//	return your key and array data
//	=====================================================================
func (j *Js) Getkey(key string, i int) *Js {
	jo := new(Js)
	num := i - 1
	if i > len((j.data).([]interface{})) {
		jo.data = errors.New("index out of range list").Error()

		return jo
	}

	if m, ok := (j.data).([]interface{}); ok {
		v := m[num].(map[string]interface{})
		if h, ok := v[key]; ok {
			jo.data = h

			return jo
		}
	}
	jo.data = nil

	return jo
}

//	=====================================================================
//	function name: Getkey
//	function type: public
//	function receiver: *Js
//	According to the custom of the PATH to find the PATH
//	=====================================================================
func (j *Js) Getpath(args ...string) *Js {
	jo := new(Js)
	d := j
	for i := range args {
		m := d.Getdata()

		if val, ok := m[args[i]]; ok {
			jo.data = val
		} else {
			jo.data = nil
			return jo
		}
	}
	return jo
}

//	=====================================================================
//	function name: Tostring
//	function type: public
//	function receiver: *Js
//	convert JSON format data to string
//	=====================================================================
func (j *Js) Tostring() string {
	if m, ok := j.data.(string); ok {
		return m
	}

	if m, ok := j.data.(float64); ok {
		return strconv.FormatFloat(m, 'f', -1, 64)
	}

	if m, ok := j.data.(bool); ok {
		return strconv.FormatBool(m)
	}

	return ""
}

//	=====================================================================
//	function name: Tostring
//	function type: public
//	function receiver: *Js
//	convert JSON format data to string array
//	=====================================================================
func (j *Js) ToArray() (k, d []string) {
	var key, data []string
	if m, ok := (j.data).([]interface{}); ok {
		for _, value := range m {
			for index, v := range value.(map[string]interface{}) {
				switch vv := v.(type) {
				case float64:
					data = append(data, strconv.FormatFloat(vv, 'f', -1, 64))
					key = append(key, index)
				case string:
					data = append(data, vv)
					key = append(key, index)

				}
			}
		}

		return key, data
	}

	if m, ok := (j.data).(map[string]interface{}); ok {
		for index, v := range m {
			switch vv := v.(type) {
			case float64:
				data = append(data, strconv.FormatFloat(vv, 'f', -1, 64))
				key = append(key, index)
			case string:
				data = append(data, vv)
				key = append(key, index)
			}
		}
		return key, data
	}

	return nil, nil
}

//	=====================================================================
//	function name: StringtoArray
//	function type: public
//	function receiver: *Js
//	convert string data to string array
//	=====================================================================
func (j *Js) StringtoArray() []string {
	var data []string
	for _, v := range j.data.([]interface{}) {
		switch vv := v.(type) {
		case string:
			data = append(data, vv)
		case float64:
			data = append(data, strconv.FormatFloat(vv, 'f', -1, 64))
		}
	}

	return data
}

//	=====================================================================
//	function name: Type
//	function type: public
//	function receiver: *Js
//	print json data type for testing
//	=====================================================================
func (j *Js) Type() {
	fmt.Println(reflect.TypeOf(j.data))
}
