package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	forceErr bool
	jsonErr  bool
}

var errForced = fmt.Errorf("this is a forced error")

func (m mockService) load(context.Context, []trainer) error {
	if m.forceErr {
		return errForced
	}

	return nil
}

func (m mockService) create(context.Context, trainer) error {
	if m.forceErr {
		return errForced
	}
	return nil
}
func (m mockService) delete(context.Context, trainer) error {
	if m.forceErr {
		return errForced
	}
	return nil
}

func (m mockService) list(context.Context) ([]*trainer, error) {
	if m.forceErr {
		return nil, errForced
	}

	trainers := []*trainer{
		{Name: "Ash", Age: 20, City: "Pallet Town"},
		{Name: "Misty", Age: 22, City: "Cerulean City"},
		{Name: "Brock", Age: 35, City: "Pewter City"},
	}

	return trainers, nil
}
func (m mockService) update(context.Context, trainer, trainer) error {
	if m.forceErr {
		return errForced
	}
	return nil
}

func TestHandlers(t *testing.T) {

	tests := map[string]struct {
		in     http.HandlerFunc
		method string
		want   string
		body   string
		status int
	}{
		"healthz": {
			in:     healthHandler,
			method: http.MethodGet,
			status: http.StatusOK,
			want:   `ok`,
		},
		"list": {
			in:     listHandler(mockService{}),
			method: http.MethodGet,
			status: http.StatusOK,
			want:   `[{"name":"Ash","age":20,"city":"Pallet Town"},{"name":"Misty","age":22,"city":"Cerulean City"},{"name":"Brock","age":35,"city":"Pewter City"}]`,
		},
		"create": {
			in:     createHandler(mockService{}),
			method: http.MethodPost,
			status: http.StatusCreated,
			body:   `{"name":"Han","age":33,"city":"Cloud City"}`,
			want:   `{"name":"Han","age":33,"city":"Cloud City"}`,
		},
		"delete": {
			in:     deleteHandler(mockService{}),
			method: http.MethodDelete,
			status: http.StatusNoContent,
			body:   `{"name":"Han","age":33,"city":"Cloud City"}`,
		},
		"update": {
			in:     updateHandler(mockService{}),
			method: http.MethodPut,
			status: http.StatusOK,
			body:   `{"original":{"name":"Han","age":33,"city":"Cloud City"},"replacement":{"name":"Han","age":33,"city":"Cloud City"}}`,
		},
		"listErr": {
			in:     listHandler(mockService{forceErr: true}),
			method: http.MethodGet,
			status: http.StatusInternalServerError,
			want:   `this is a forced error`,
		},
		"createErr": {
			in:     createHandler(mockService{forceErr: true}),
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			body:   `{"name":"Han","age":33,"city":"Cloud City"}`,
			want:   `this is a forced error`,
		},
		"deleteErr": {
			in:     deleteHandler(mockService{forceErr: true}),
			method: http.MethodDelete,
			status: http.StatusInternalServerError,
			body:   `{"name":"Han","age":33,"city":"Cloud City"}`,
			want:   `this is a forced error`,
		},
		"updateErr": {
			in:     updateHandler(mockService{forceErr: true}),
			method: http.MethodPut,
			status: http.StatusInternalServerError,
			body:   `{"original":{"name":"Han","age":33,"city":"Cloud City"},"replacement":{"name":"Han","age":33,"city":"Cloud City"}}`,
			want:   `this is a forced error`,
		},
		"createJSONErr": {
			in:     createHandler(mockService{forceErr: true}),
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			body:   `this aint no json`,
			want:   `invalid character 'h' in literal true (expecting 'r')`,
		},
		"deleteJSONErr": {
			in:     deleteHandler(mockService{forceErr: true}),
			method: http.MethodDelete,
			status: http.StatusInternalServerError,
			body:   `this aint no json`,
			want:   "invalid character 'h' in literal true (expecting 'r')",
		},
		"updateJSONErr": {
			in:     updateHandler(mockService{forceErr: true}),
			method: http.MethodPut,
			status: http.StatusInternalServerError,
			body:   `this aint no json`,
			want:   `invalid character 'h' in literal true (expecting 'r')`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			body := []byte(tc.body)

			req := httptest.NewRequest(tc.method, "/api/v1/trainer", bytes.NewReader(body))
			w := httptest.NewRecorder()
			tc.in(w, req)
			res := w.Result()
			defer res.Body.Close()
			got, err := ioutil.ReadAll(res.Body)

			require.Nil(t, err)
			assert.Equal(t, tc.want, string(got))

		})
	}
}
