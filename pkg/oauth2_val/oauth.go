package oauth2_val

import (
	"context"
	"log"
	"net/http"
	"oauth2/config"
	"oauth2/pkg/model"
	"oauth2/pkg/session"
	"oauth2/pkg/storage"
	"strconv"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt/v5"
)

// Srv 是 OAuth2 Server 的核心实例，负责处理授权码、令牌等 HTTP 请求。
var Srv *server.Server

// Mgr 是 OAuth2 的管理器，负责令牌存储、客户端信息、Token 配置等资源管理。
var Mgr *manage.Manager

func Setup(ctx context.Context) {
	// 创建默认管理器，负责 token 管理、客户端存储、配置等
	Mgr = manage.NewDefaultManager()
	// 设置授权码模式下 AccessToken/RefreshToken 的生成策略
	Mgr.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * time.Duration(config.GetCfg().OAuth2.AccessTokenExp),
		RefreshTokenExp:   time.Hour * 24 * 3,
		IsGenerateRefresh: true,
	})
	switch config.GetCfg().OAuth2.TokenStore {
	case "memory":
		Mgr.MustTokenStorage(store.NewMemoryTokenStore())
	case "redis":
		// Mgr.MustTokenStorage(store.NewRedisTokenStore())
	case "mysql":
		sqlDb, err := model.GlobalDB.DB()
		if err != nil {
			panic(err)
		}
		tokenStore := storage.NewMySQLTokenStore(sqlDb, "access_tokens")
		tokenStore.StartTicker(ctx, 5*time.Second)
		// 创建表
		if err := tokenStore.CreateTable(); err != nil {
			log.Fatal("Failed to create token table:", err)
		}
		Mgr.MustTokenStorage(tokenStore, nil)
	default:
		Mgr.MustTokenStorage(store.NewMemoryTokenStore())
	}
	// 配置 JWT Access Token 的生成器
	Mgr.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(config.GetCfg().OAuth2.JWTSignedKey), jwt.SigningMethodHS512))
	// 注册 Client 信息（可以改成 DB/配置中心）
	clientStore := store.NewClientStore()
	for _, v := range config.GetCfg().OAuth2.Client {
		err := clientStore.Set(v.ID, &models.Client{
			ID:     v.ID,
			Secret: v.Secret,
			Domain: v.Domain,
		})
		if err != nil {
			return
		}
	}
	Mgr.MapClientStorage(clientStore)

	// 创建 OAuth2 Server 实例并挂载各类 Handler
	Srv = server.NewServer(server.NewConfig(), Mgr)
	Srv.SetPasswordAuthorizationHandler(passwordAuthorizationHandler) // 处理 “password” 授权模式（资源所有者密码凭证）时的用户验证逻辑，当客户端提交用户名 + 密码换取 token 时调用。
	Srv.SetUserAuthorizationHandler(userAuthorizeHandler)             // 处理 “authorization_code” 等需要用户确认授权的流程，用来检查当前是否已有登录用户；如果没有，通常重定向到登录页
	Srv.SetAuthorizeScopeHandler(authorizeScopeHandler)               // 当用户勾选/确认授权范围（scope）后，对比客户端注册的合法 scope，过滤非法项，并返回最终生效的 scope
	Srv.SetInternalErrorHandler(internalErrorHandler)                 // OAuth2 server 内部出错（例如存储、生成 token 时异常）时的统一兜底处理，可以记录日志、定制返回
	Srv.SetResponseErrorHandler(responseErrorHandler)                 // 当 OAuth2 协议对外响应发生错误（如无效客户端、无效授权）时的处理，可用于统一日志或格式化错误输出
}

// oauth2进行密码认证的方式
func passwordAuthorizationHandler(ctx context.Context, clientID, username, password string) (userID string, err error) {
	var user model.User
	userIDInt, err1 := user.Authentication(ctx, clientID, username, password)
	userID = strconv.Itoa(int(userIDInt))
	err = err1
	return
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	v, _ := session.Get(r, "LoggedInUserID")
	if v == nil {
		if r.Form == nil {
			r.ParseForm()
		}
		session.Set(w, r, "RequestForm", r.Form)

		// 登录页面
		// 最终会把userId写进session(LoggedInUserID)
		// 再跳回来
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	userID = v.(string)
	return
}

// 场景:在登录页面勾选所要访问的资源范围
// 根据client注册的scope,过滤表单中非法scope
// HandleAuthorizeRequest中调用
// set scope for the access token
func authorizeScopeHandler(w http.ResponseWriter, r *http.Request) (scope string, err error) {
	if r.Form == nil {
		r.ParseForm()
	}
	s := config.ScopeFilter(r.Form.Get("client_id"), r.Form.Get("scope"))
	if s == nil {
		err = errors.New("无效的权限范围")
		return
	}
	scope = config.JoinScope(s)
	return
}

func internalErrorHandler(err error) (re *errors.Response) {
	log.Println("Internal Error:", err.Error())
	return
}

func responseErrorHandler(re *errors.Response) {
	log.Println("Response Error:", re.Error.Error())
}
