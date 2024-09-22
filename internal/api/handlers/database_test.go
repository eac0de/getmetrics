package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eac0de/getmetrics/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	database := mocks.NewMockIDatabase(ctrl)
	database.EXPECT().PingContext(gomock.Any()).Return(nil)
	dh := NewDatabaseHandlers(database)
	t.Run("200 ping", func(t *testing.T) {
		url := "/ping"
		r := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		dh.PingHandler()(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	database.EXPECT().PingContext(gomock.Any()).Return(fmt.Errorf(""))
	t.Run("500 ping", func(t *testing.T) {
		url := "/ping"
		r := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		dh.PingHandler()(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
