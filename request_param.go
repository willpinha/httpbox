package httpbox

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type paramFrom string

const (
	fromPath  paramFrom = "URL path"
	fromQuery paramFrom = "URL query string"
)

type Param struct {
	from  paramFrom
	name  string
	value string
}

func (p Param) newError(msg string) error {
	return NewError(
		http.StatusBadRequest,
		fmt.Sprintf("parameter %q from %s %s", p.name, p.from, msg),
	)
}

func NewPathParam(r *http.Request, name string) Param {
	return Param{
		from:  fromPath,
		name:  name,
		value: r.PathValue(name),
	}
}

func NewQueryParam(r *http.Request, name string) Param {
	return Param{
		from:  fromQuery,
		name:  name,
		value: r.URL.Query().Get(name),
	}
}

func NewDefaultQueryParam(r *http.Request, name, defaultValue string) Param {
	p := NewQueryParam(r, name)

	if p.value == "" {
		p.value = defaultValue
	}

	return p
}

func NewRequiredQueryParam(r *http.Request, name string) (Param, error) {
	p := NewQueryParam(r, name)

	if p.value == "" {
		return Param{}, p.newError("is required")
	}

	return p, nil
}

func (p Param) String() string {
	return p.value
}

func (p Param) Int() (int, error) {
	v, err := strconv.Atoi(p.value)
	if err != nil {
		return 0, p.newError("must be an integer")
	}

	return v, nil
}

func (p Param) Float() (float64, error) {
	v, err := strconv.ParseFloat(p.value, 64)
	if err != nil {
		return 0, p.newError("must be a float. Example value: 3.14")
	}

	return v, nil
}

func (p Param) Bool() (bool, error) {
	v, err := strconv.ParseBool(p.value)
	if err != nil {
		return false, p.newError("must be a boolean. Example values: true, false, 1, 0")
	}

	return v, nil
}

func (p Param) Time(format string) (time.Time, error) {
	v, err := time.Parse(format, p.value)
	if err != nil {
		exampleValue := time.Now().Format(format)

		return time.Time{}, p.newError(fmt.Sprintf("must be a time. Example value: %s", exampleValue))
	}

	return v, nil
}

func (p Param) Duration() (time.Duration, error) {
	v, err := time.ParseDuration(p.value)
	if err != nil {
		return 0, p.newError("must be a time duration. Example values: 300ms, -1.5h, 2h45m")
	}

	return v, nil
}
