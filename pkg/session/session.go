package session

import (
	"encoding/gob"
	"github.com/gorilla/sessions"
	"net/http"
	"net/url"
	"oauth2/config"
)

var store *sessions.CookieStore

func Setup() {
	gob.Register(url.Values{})

	store = sessions.NewCookieStore([]byte(config.GetCfg().Session.SecretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   config.GetCfg().Session.MaxAge,
		HttpOnly: true,
	}
}

// Get 获得session中key对应的值
func Get(r *http.Request, name string) (val interface{}, err error) {
	session, err := store.Get(r, config.GetCfg().Session.Name)
	if err != nil {
		return
	}
	val = session.Values[name]
	return
}

// Set 设置session里的键值对
func Set(w http.ResponseWriter, r *http.Request, name string, val interface{}) (err error) {
	session, err := store.Get(r, config.GetCfg().Session.Name)
	if err != nil {
		return
	}
	session.Values[name] = val
	err = sessions.Save(r, w)
	return
}

func Delete(w http.ResponseWriter, r *http.Request, name string) (err error) {
	session, err := store.Get(r, config.GetCfg().Session.Name)
	if err != nil {
		return
	}
	delete(session.Values, name)
	err = sessions.Save(r, w)
	return
}
