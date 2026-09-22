package router

import (
	"fmt"
	"net/http/httptest"
	"one-api/common/config"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserIndexConcurrentConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.OptionMapRWMutex.Lock()
	original := config.OptionMap
	config.OptionMap = map[string]string{"SystemText": ""}
	config.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		config.OptionMapRWMutex.Lock()
		config.OptionMap = original
		config.OptionMapRWMutex.Unlock()
	})
	page := []byte(`<html><script src="/static/js/main.abcdef.js"></script></html>`)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			config.OptionMapRWMutex.Lock()
			config.OptionMap["SystemText"] = fmt.Sprintf(`<html>%d<script src="/static/js/main.123456.js"></script></html>`, i)
			config.OptionMapRWMutex.Unlock()
		}
	}()
	for n := 0; n < 32; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Set("defaultUserIndexPage", page)
				serveUserIndexPage(c)
				if w.Code != 200 || !strings.Contains(w.Body.String(), "main.abcdef.js") {
					t.Errorf("unexpected index response: %d", w.Code)
					return
				}
			}
		}()
	}
	wg.Wait()
	if !strings.Contains(config.GetOption("SystemText"), "main.123456.js") {
		t.Fatal("serving the page modified stored configuration")
	}
}
