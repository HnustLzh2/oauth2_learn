package oauth2_val

import (
	"context"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"oauth2/config"
	"oauth2/pkg/model"
	"oauth2/pkg/session"
	"strconv"
	"time"
)

var Srv *server.Server
var Mgr *manage.Manager

func Setup() {
	Mgr = manage.NewDefaultManager()
	Mgr.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * time.Duration(config.GetCfg().OAuth2.AccessTokenExp),
		RefreshTokenExp:   time.Hour * 24 * 3,
		IsGenerateRefresh: true,
	})
	Mgr.MustTokenStorage(store.NewMemoryTokenStore())
	Mgr.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(config.GetCfg().OAuth2.JWTSignedKey), jwt.SigningMethodHS512))
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

	Srv = server.NewServer(server.NewConfig(), Mgr)
	Srv.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)
	Srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	Srv.SetAuthorizeScopeHandler(authorizeScopeHandler)
	Srv.SetInternalErrorHandler(internalErrorHandler)
	Srv.SetResponseErrorHandler(responseErrorHandler)
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
