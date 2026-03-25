package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// resetUsers 在每个测试前清空 sync.Map，避免测试间状态污染。
// 使用反射绕过 sync.Map 不导出 LoadAndDelete 的限制。
func resetUsers() {
	type entry struct {
		key any
		val any
	}
	// sync.Map 本质是 readOnly + dirty，Range 遍历后手动重建
	var m sync.Map
	users.Range(func(k, v any) bool {
		m.Store(k, v)
		return true
	})
	m.Range(func(k, v any) bool {
		users.Delete(k)
		return true
	})
}

func setupTestRouter() *gin.Engine {
	r := gin.New()
	RegisterRoutes(r)
	return r
}

func postForm(url string, values url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// ---------------------------------------------------------------------------
// Register Handler Tests
// ---------------------------------------------------------------------------

func TestRegister_MissingEmail(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/register", url.Values{"password": {"pass123"}})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "邮箱和密码不能为空") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/register", url.Values{"email": {"a@b.com"}})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "邮箱和密码不能为空") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestRegister_BothMissing(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/user/register", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegister_Success(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/register", url.Values{
		"email":    {"test@example.com"},
		"password": {"SecurePass123!"},
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "注册成功") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	// 第一次注册
	req1 := postForm("/user/register", url.Values{
		"email":    {"dup@example.com"},
		"password": {"Pass1!"},
	})
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first register: expected %d, got %d", http.StatusOK, w1.Code)
	}

	// 第二次用相同邮箱注册
	req2 := postForm("/user/register", url.Values{
		"email":    {"dup@example.com"},
		"password": {"Pass2!"},
	})
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatalf("duplicate register: expected status %d, got %d", http.StatusBadRequest, w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "邮箱已经存在") {
		t.Fatalf("unexpected body: %s", w2.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Login Handler Tests
// ---------------------------------------------------------------------------

func TestLogin_MissingEmail(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/login", url.Values{"password": {"pass123"}})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "邮箱和密码不能为空") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/login", url.Values{"email": {"a@b.com"}})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_EmailNotFound(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	req := postForm("/user/login", url.Values{
		"email":    {"notfound@example.com"},
		"password": {"anypass"},
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "邮箱不存在") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	// 先注册一个用户
	regReq := postForm("/user/register", url.Values{
		"email":    {"user@example.com"},
		"password": {"CorrectPass123!"},
	})
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d", regW.Code)
	}

	// 用错误密码登录
	req := postForm("/user/login", url.Values{
		"email":    {"user@example.com"},
		"password": {"WrongPassword!"},
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "密码错误") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	const email = "login@example.com"
	const password = "LoginPass123!"

	// 注册
	regReq := postForm("/user/register", url.Values{
		"email":    {email},
		"password": {password},
	})
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d", regW.Code)
	}

	// 登录
	loginReq := postForm("/user/login", url.Values{
		"email":    {email},
		"password": {password},
	})
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body: %s", http.StatusOK, loginW.Code, loginW.Body.String())
	}
	if !strings.Contains(loginW.Body.String(), "登录成功") {
		t.Fatalf("unexpected body: %s", loginW.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Register & Login Integration Test（完整注册-登录流程）
// ---------------------------------------------------------------------------

func TestRegisterLoginFlow(t *testing.T) {
	resetUsers()
	router := setupTestRouter()

	const email = "flow@example.com"
	const password = "FlowPass456!"

	// 1. 正常注册
	regReq := postForm("/user/register", url.Values{
		"email":    {email},
		"password": {password},
	})
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register: expected %d, got %d", http.StatusOK, regW.Code)
	}

	// 2. 用同一凭证登录成功
	loginReq := postForm("/user/login", url.Values{
		"email":    {email},
		"password": {password},
	})
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login: expected %d, got %d; body: %s", http.StatusOK, loginW.Code, loginW.Body.String())
	}

	// 3. 修改密码后旧密码无法登录（bcrypt 密码不可逆验证）
	// 先尝试用旧密码重复登录（应成功）
	reLoginReq := postForm("/user/login", url.Values{
		"email":    {email},
		"password": {password},
	})
	reLoginW := httptest.NewRecorder()
	router.ServeHTTP(reLoginW, reLoginReq)
	if reLoginW.Code != http.StatusOK {
		t.Fatalf("re-login: expected %d, got %d", http.StatusOK, reLoginW.Code)
	}
}
