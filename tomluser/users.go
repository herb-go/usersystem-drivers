package tomluser

import (
	"sort"
	"sync"

	"github.com/herb-go/herbsecurity/authorize/role"
	"github.com/herb-go/uniqueid"

	"github.com/herb-go/providers/herb/statictoml"

	"github.com/herb-go/user"
	"github.com/herb-go/user/profile"
	"github.com/herb-go/user/status"
)

type Users struct {
	Source     statictoml.Source
	locker     sync.RWMutex
	uidmap     map[string]*User
	accountmap map[string][]*User
	idFactory  func() (string, error)
	HashMode   string
	status.Service
	ProfileFields map[string]bool
}

func (u *Users) GetProfile(id string) (*profile.Profile, error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	userdata := u.uidmap[id]
	if userdata == nil {
		return nil, user.ErrUserNotExists
	}
	return userdata.Profiles.Clone(), nil
}
func (u *Users) UpdateProfile(id string, p *profile.Profile) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	userdata := u.uidmap[id]
	if userdata == nil {
		return user.ErrUserNotExists
	}
	prf := profile.NewProfile()
	for _, v := range p.Data() {
		if u.ProfileFields[v.Name] {
			prf.WithFields(v)
		}
	}
	userdata.Profiles = prf
	return u.save()
}

func (u *Users) GetAllUsers() *Data {
	data := NewData()
	data.Users = make([]*User, 0, len(u.uidmap))
	for k := range u.uidmap {
		data.Users = append(data.Users, u.uidmap[k])
	}
	return data
}
func (u *Users) save() error {
	return u.Source.Save(u.GetAllUsers())
}
func (u *Users) LoadStatus(id string) (status.Status, error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	var st status.Status
	userdata := u.uidmap[id]
	if userdata == nil {
		return st, user.ErrUserNotExists
	}
	if userdata.Banned {
		st = status.StatusBanned
	} else {
		st = status.StatusNormal
	}
	return st, nil

}
func (u *Users) UpdateStatus(uid string, st status.Status) error {
	u.locker.Lock()
	defer u.locker.Unlock()

	if u.uidmap[uid] == nil {
		return user.ErrUserNotExists
	}
	ok, err := status.NormalOrBannedService.IsAvailable(st)
	if err != nil {
		return err
	}
	u.uidmap[uid].Banned = !ok
	return u.save()

}
func (u *Users) CreateStatus(uid string) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	if u.uidmap[uid] != nil {
		return user.ErrUserExists
	}
	newuser := NewUser()
	newuser.UID = uid
	u.addUser(newuser)
	return u.save()
}
func (u *Users) RemoveStatus(uid string) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	if u.uidmap[uid] == nil {
		return user.ErrUserNotExists
	}
	u.removeUser(uid)
	return u.save()

}
func (u *Users) getAfterLast(last string, users []string) []string {
	sort.Strings(users)
	if last == "" {
		return users
	}
	for k := range users {
		if users[k] > last {
			return users[k:]
		}
	}
	return []string{}
}

func (u *Users) ListUsersByStatus(last string, limit int, statuses ...status.Status) ([]string, bool, error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	m := map[bool]bool{}
	for k := range statuses {
		ok, err := u.IsAvailable(statuses[k])
		if err != nil {
			return nil, false, nil
		}
		m[!ok] = true
	}
	users := []string{}
	for k := range u.uidmap {
		if m[u.uidmap[k].Banned] {
			users = append(users, k)
		}
	}
	result := u.getAfterLast(last, users)
	if limit > 0 && limit < len(result) {
		return result[:limit], false, nil
	}
	return result, true, nil
}

//VerifyPassword Verify user password.
//Return verify result and any error if raised
func (u *Users) VerifyPassword(uid string, password string) (bool, error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	user := u.uidmap[uid]
	if user == nil {
		return false, nil
	}
	return user.VerifyPassword(password)
}

//PasswordChangeable return password changeable
func (u *Users) PasswordChangeable() bool {
	return true
}

//UpdatePassword update user password
//Return any error if raised
func (u *Users) UpdatePassword(uid string, password string) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		return user.ErrUserNotExists
	}
	err := us.UpdatePassword(u.HashMode, password)
	if err != nil {
		return err
	}
	return u.save()
}
func (u *Users) Roles(uid string) (*role.Roles, error) {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		return nil, user.ErrUserNotExists
	}
	return us.Roles.Clone(), nil
}
func (u *Users) SetRoles(uid string, r *role.Roles) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		return user.ErrUserNotExists
	}
	us.Roles = r
	return nil
}
func (u *Users) Account(uid string) (*user.Accounts, error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	us := u.uidmap[uid]
	if us == nil {
		return nil, user.ErrUserNotExists
	}
	accs := user.Accounts(us.Accounts)
	return &accs, nil
}

func (u *Users) accountToUID(account *user.Account) (uid string, err error) {
	for _, user := range u.accountmap[account.Account] {
		for k := range user.Accounts {
			if user.Accounts[k].Equal(account) {
				return user.UID, nil
			}
		}
	}
	return "", nil
}

//AccountToUID query uid by user account.
//Return user id and anyidFactory error if raised.
//Return empty string as userid if account not found.
func (u *Users) AccountToUID(account *user.Account) (uid string, err error) {
	u.locker.RLock()
	defer u.locker.RUnlock()
	return u.accountToUID(account)
}

//BindAccount bind account to user.
//Return any error if raised.
//If account exists,user.ErrAccountBindingExists should be rasied.
func (u *Users) BindAccount(uid string, account *user.Account) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	accountuser := u.uidmap[uid]
	if accountuser == nil {
		return user.ErrUserNotExists
	}
	accountid, err := u.accountToUID(account)
	if err != nil {
		return err
	}
	if accountid != "" {
		return user.ErrAccountBindingExists
	}
	accountuser.Accounts = append(accountuser.Accounts, account)
	u.accountmap[account.Account] = append(u.accountmap[account.Account], accountuser)
	return u.save()
}

//UnbindAccount unbind account from user.
//Return any error if raised.
//If account not exists,user.ErrAccountUnbindingNotExists should be rasied.
func (u *Users) UnbindAccount(uid string, account *user.Account) error {
	u.locker.Lock()
	defer u.locker.Unlock()
	accountuser := u.uidmap[uid]
	if accountuser == nil {
		return user.ErrUserNotExists
	}
	accountid, err := u.accountToUID(account)
	if err != nil {
		return err
	}
	if accountid == "" || accountid != uid {
		return user.ErrAccountUnbindingNotExists
	}
	for k := range u.uidmap[accountid].Accounts {
		if u.uidmap[accountid].Accounts[k].Equal(account) {
			u.uidmap[accountid].Accounts = append(u.uidmap[accountid].Accounts[:k], u.uidmap[accountid].Accounts[k+1:]...)
			break
		}
	}
	for k := range u.accountmap[account.Account] {
		if u.accountmap[account.Account][k].UID == accountid {
			u.accountmap[account.Account] = append(u.accountmap[account.Account][:k], u.accountmap[account.Account][k+1:]...)
			break
		}
	}
	return u.save()
}

func (u *Users) CurrentTerm(uid string) (string, error) {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		return "", user.ErrUserNotExists
	}
	return us.Term, nil

}
func (u *Users) StartNewTerm(uid string) (string, error) {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		return "", user.ErrUserNotExists
	}
	term, err := u.idFactory()
	if err != nil {
		return "", err
	}
	us.Term = term
	return term, nil
}

func (u *Users) addUser(user *User) {
	u.uidmap[user.UID] = user
	for _, a := range user.Accounts {
		u.accountmap[a.Account] = append(u.accountmap[a.Keyword], user)
	}
}

func (u *Users) removeUser(uid string) {
	user := u.uidmap[uid]
	delete(u.uidmap, uid)
	for _, a := range user.Accounts {
		accounts := make([]*User, 0, len(u.accountmap[a.Account])-1)
		for _, v := range u.accountmap[a.Account] {
			accounts = append(accounts, v)
		}
		u.accountmap[a.Account] = accounts
	}
}

func (u *Users) Purge(uid string) error {
	return nil
}
func (u *Users) Start() error {
	return nil
}
func (u *Users) Stop() error {
	return nil
}

func NewUsers() *Users {
	return &Users{
		uidmap:        map[string]*User{},
		accountmap:    map[string][]*User{},
		idFactory:     uniqueid.DefaultGenerator.GenerateID,
		HashMode:      defaultUsersHashMode,
		Service:       status.NormalOrBannedService,
		ProfileFields: map[string]bool{},
	}
}
