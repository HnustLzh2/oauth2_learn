package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"oauth2/config"
	"oauth2/pkg/model"
	"oauth2/pkg/oauth2_val"
	"oauth2/pkg/session"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// getProjectRoot 获取项目根目录（go.mod所在目录）
func getProjectRoot() string {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: failed to get working directory: %v", err)
		return "."
	}

	// 尝试从当前目录向上查找 go.mod 文件来确定项目根目录
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// 找到 go.mod，说明这是项目根目录
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达文件系统根目录，使用当前工作目录
			break
		}
		dir = parent
	}

	// 如果找不到 go.mod，使用当前工作目录
	return wd
}

// GetTemplatePath 获取模板文件的绝对路径
// 基于项目根目录（go.mod所在目录）来解析路径
func GetTemplatePath(filename string) string {
	return filepath.Join(getProjectRoot(), filename)
}

// 错误显示页面
// 以网页的形式展示大于400的错误
func errorHandler(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	if status >= 400 {
		t, _ := template.ParseFiles(GetTemplatePath("tpl/error.html"))
		body := struct {
			Status  int
			Message string
		}{Status: status, Message: message}
		t.Execute(w, body)
	}
}
func AuthorizeHandler(ctx *gin.Context) {
	var form url.Values
	if v, _ := session.Get(ctx.Request, "RequestForm"); v != nil {
		ctx.Request.ParseForm()
		if ctx.Request.Form.Get("client_id") == "" {
			form = v.(url.Values)
		}
	}
	ctx.Request.Form = form

	if err := session.Delete(ctx.Writer, ctx.Request, "RequestForm"); err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := oauth2_val.Srv.HandleAuthorizeRequest(ctx.Writer, ctx.Request); err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusBadRequest)
		return
	}
}

type TplData struct {
	Client config.OAuth2Client
	// 用户申请合规的scope
	Scope []config.Scope
	Error string
}

func LoginHandler(ctx *gin.Context) {
	form, _ := session.Get(ctx.Request, "RequestForm")
	if form == nil {
		errorHandler(ctx.Writer, "无效的请求", http.StatusInternalServerError)
		return
	}
	clientID := form.(url.Values).Get("client_id")
	scope := form.(url.Values).Get("scope")

	data := TplData{
		Client: *config.GetOAuth2Client(clientID),
		Scope:  config.ScopeFilter(clientID, scope),
	}
	if data.Scope == nil {
		errorHandler(ctx.Writer, "无效的权限范围", http.StatusBadRequest)
		return
	}
	var userID string
	var err error
	if ctx.Request.Form == nil {
		err = ctx.Request.ParseForm()
		if err != nil {
			errorHandler(ctx.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// 进行登入验证
	if ctx.Request.Form.Get("type") == "password" {
		var user model.User
		userIDUint, err1 := user.Authentication(ctx, ctx.Request.Form.Get("client_id"), ctx.Request.Form.Get("username"), ctx.Request.Form.Get("password"))
		userID = strconv.Itoa(int(userIDUint))
		err = err1
		if err != nil {
			data.Error = err.Error()
			t, _ := template.ParseFiles(GetTemplatePath("tpl/login.html"))
			t.Execute(ctx.Writer, data)
			return
		}
	}
	// 可以进行其他方式的验证
	err = session.Set(ctx.Writer, ctx.Request, "LoggedInUserID", userID)
	if err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.Writer.Header().Set("Location", "/authorize")
	ctx.Writer.WriteHeader(http.StatusFound)
}

func GETloginHandler(ctx *gin.Context) {
	form, _ := session.Get(ctx.Request, "RequestForm")
	if form == nil {
		errorHandler(ctx.Writer, "无效的请求", http.StatusInternalServerError)
		return
	}
	clientID := form.(url.Values).Get("client_id")
	scope := form.(url.Values).Get("scope")

	data := TplData{
		Client: *config.GetOAuth2Client(clientID),
		Scope:  config.ScopeFilter(clientID, scope),
	}
	if data.Scope == nil {
		errorHandler(ctx.Writer, "无效的权限范围", http.StatusBadRequest)
		return
	}
	// 如果为GET方法就直接返回登录页面
	t, err := template.ParseFiles(GetTemplatePath("tpl/login.html"))
	if err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(ctx.Writer, data); err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func LogoutHandler(ctx *gin.Context) {
	if ctx.Request.Form == nil {
		if err := ctx.Request.ParseForm(); err != nil {
			errorHandler(ctx.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// 检查redirect_uri参数
	redirectURI := ctx.Request.Form.Get("redirect_uri")
	if redirectURI == "" {
		errorHandler(ctx.Writer, "参数不能为空(redirect_uri)", http.StatusBadRequest)
		return
	}
	if _, err := url.Parse(redirectURI); err != nil {
		errorHandler(ctx.Writer, "参数无效(redirect_uri)", http.StatusBadRequest)
		return
	}

	// 删除公共回话
	if err := session.Delete(ctx.Writer, ctx.Request, "LoggedInUserID"); err != nil {
		errorHandler(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.Writer.Header().Set("Location", redirectURI)
	ctx.Writer.WriteHeader(http.StatusFound)
}

func TokenHandler(ctx *gin.Context) {
	err := oauth2_val.Srv.HandleTokenRequest(ctx.Writer, ctx.Request)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func VerifyHandler(ctx *gin.Context) {
	token, err := oauth2_val.Srv.ValidationBearerToken(ctx.Request)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusUnauthorized)
		return
	}
	cli, err := oauth2_val.Mgr.GetClient(ctx.Request.Context(), token.GetClientID())
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusBadRequest)
		return
	}
	data := map[string]interface{}{
		"expires_in": int64(time.Until(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn())).Seconds()),
		"user_id":    token.GetUserID(),
		"client_id":  token.GetClientID(),
		"scope":      token.GetScope(),
		"domain":     cli.GetDomain(),
	}
	e := json.NewEncoder(ctx.Writer)
	e.SetIndent("", "  ")
	e.Encode(data)
}

func NotFoundHandler(ctx *gin.Context) {
	errorHandler(ctx.Writer, "无效的地址", http.StatusNotFound)
}
