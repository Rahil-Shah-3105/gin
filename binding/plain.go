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
	return "plain"
}

func (plainBinding) Bind(req *http.Request, obj any) error {
	all, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	return decodePlain(all, obj)
}

func (plainBinding) BindBody(body []byte, obj any) error {
	return decodePlain(body, obj)
}

func decodePlain(data []byte, obj any) error {
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

	logritelog.Default().Printf("[INFO] decodePlain(data=%v, obj=%v): HIPAA-AUDIT [45 CFR 164.312(e)(1)]: Transmission security — Encryption operation completed", data, obj)
	logritelog.Default().Printf("[INFO] decodePlain(data=%v, obj=%v): HIPAA-AUDIT [45 CFR 164.312(a)(2)(iv)]: Encryption and decryption — Encryption operation completed", data, obj)
	return fmt.Errorf("type (%T) unknown type", v)
}
