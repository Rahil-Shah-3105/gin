package binding
import logritelog "log"

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin/internal/bytesconv"
)

type plainBinding struct{}

func (plainBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	return "plain"
}

func (plainBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	all, err := io.ReadAll(req.Body)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Bind(req=%v,obj=%v): %v", req, obj, err)
		return err
	}

	logritelog.Default().Printf("[WARN] <<< Exiting Bind(req=%v,obj=%v)", req, obj)
	return decodePlain(all, obj)
}

func (plainBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	return decodePlain(body, obj)
}

func decodePlain(data []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering decodePlain(data=%v,obj=%v)", data, obj)
	if obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.String {
		v.SetString(bytesconv.BytesToString(data))
		return nil
	}

	if _, ok := v.Interface().([]byte); ok {
		v.SetBytes(data)
		return nil
	}

	logritelog.Default().Printf("[WARN] <<< Exiting decodePlain(data=%v,obj=%v)", data, obj)
	return fmt.Errorf("type (%T) unknown type", v)
}
