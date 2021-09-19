package store

import (
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"sync"
	"time"
)

type user struct {
	Lock    sync.RWMutex
	File    string
	Users   *libol.SafeStrMap
	LdapCfg *libol.LDAPConfig
	LdapSvc *libol.LDAPService
}

func (w *user) Save() error {
	if w.File == "" {
		return nil
	}
	fp, err := libol.OpenTrunk(w.File)
	if err != nil {
		return err
	}
	for obj := range w.List() {
		if obj == nil {
			break
		}
		if obj.Role == "ldap" {
			continue
		}
		line := obj.Id()
		line += ":" + obj.Password
		line += ":" + obj.Role
		line += ":" + obj.Lease.Format(libol.LeaseTime)
		_, _ = fp.WriteString(line + "\n")
	}
	return nil
}

func (w *user) SetFile(value string) {
	w.File = value
}

func (w *user) Init(size int) {
	w.Users = libol.NewSafeStrMap(size)
}

func (w *user) Add(user *models.User) {
	libol.Debug("user.Add %v", user)
	key := user.Id()
	if older := w.Get(key); older == nil {
		_ = w.Users.Set(key, user)
	} else { // Update pass and role.
		older.Role = user.Role
		older.Password = user.Password
		older.Alias = user.Alias
		older.UpdateAt = user.UpdateAt
		older.Lease = user.Lease
	}
}

func (w *user) Del(key string) {
	libol.Debug("user.Add %s", key)
	w.Users.Del(key)
}

func (w *user) Get(key string) *models.User {
	if v := w.Users.Get(key); v != nil {
		return v.(*models.User)
	}
	return nil
}

func (w *user) List() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.Users.Iter(func(k string, v interface{}) {
			c <- v.(*models.User)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *user) CheckLdap(obj *models.User) *models.User {
	svc := w.GetLdap()
	if svc == nil {
		return nil
	}
	u := w.Get(obj.Id())
	libol.Debug("CheckLdap %s", u)
	if u != nil && u.Role != "ldap" {
		return nil
	}
	if ok, err := svc.Login(obj.Id(), obj.Password); !ok {
		libol.Warn("CheckLdap %s", err)
		return nil
	}
	user := &models.User{
		Name:     obj.Id(),
		Password: obj.Password,
		Role:     "ldap",
		Alias:    obj.Alias,
	}
	user.Update()
	w.Add(user)
	return user
}

func (w *user) Timeout(user *models.User) bool {
	if user.Role == "ldap" {
		return time.Now().Unix()-user.UpdateAt > w.LdapCfg.Timeout
	}
	return true
}

func (w *user) Check(obj *models.User) *models.User {
	if u := w.Get(obj.Id()); u != nil {
		if u.Role == "" || u.Role == "admin" || u.Role == "guest" {
			if u.Password == obj.Password {
				t0 := time.Now()
				t1 := u.Lease
				if t1.Year() < 2000 || t1.After(t0) {
					return u
				}
				return nil
			}
		}
	}
	if u := w.CheckLdap(obj); u != nil {
		return u
	}
	return nil
}

func (w *user) GetLdap() *libol.LDAPService {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if w.LdapCfg == nil {
		return nil
	}
	if w.LdapSvc == nil || w.LdapSvc.Conn.IsClosing() {
		if l, err := libol.NewLDAPService(*w.LdapCfg); err != nil {
			libol.Warn("user.GetLdap %s", err)
			w.LdapSvc = nil
		} else {
			w.LdapSvc = l
		}
	}
	return w.LdapSvc
}

func (w *user) SetLdap(cfg *libol.LDAPConfig) {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if w.LdapCfg != cfg {
		w.LdapCfg = cfg
	}
	if l, err := libol.NewLDAPService(*cfg); err != nil {
		libol.Warn("user.SetLdap %s", err)
	} else {
		libol.Info("user.SetLdap %s", w.LdapCfg.Server)
		w.LdapSvc = l
	}
}

var User = user{
	Users: libol.NewSafeStrMap(1024),
}
