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

func (u *Users) MustGetProfile(id string) *profile.Profile {
	u.locker.RLock()
	defer u.locker.RUnlock()
	userdata := u.uidmap[id]
	if userdata == nil {
		panic(user.ErrUserNotExists)
	}
	return userdata.Profiles.Clone()
}
func (u *Users) MustUpdateProfile(id string, p *profile.Profile) {
	u.locker.Lock()
	defer u.locker.Unlock()
	userdata := u.uidmap[id]
	if userdata == nil {
		panic(user.ErrUserNotExists)
	}
	prf := profile.NewProfile()
	for _, v := range p.Data() {
		if u.ProfileFields[v.Name] {
			prf.WithFields(v)
		}
	}
	userdata.Profiles = prf
	u.mustSave()
}

func (u *Users) GetAllUsers() *Data {
	data := NewData()
	data.Users = make([]*User, 0, len(u.uidmap))
	for k := range u.uidmap {
		data.Users = append(data.Users, u.uidmap[k])
	}
	return data
}
func (u *Users) mustSave() {
	err := u.save()
	if err != nil {
		panic(err)
	}
}

func (u *Users) save() error {
	return u.Source.Save(u.GetAllUsers())
}
func (u *Users) MustLoadStatus(id string) status.Status {
	u.locker.RLock()
	defer u.locker.RUnlock()
	var st status.Status
	userdata := u.uidmap[id]
	if userdata == nil {
		panic(user.ErrUserNotExists)
	}
	if userdata.Banned {
		st = status.StatusBanned
	} else {
		st = status.StatusNormal
	}
	return st

}
func (u *Users) MustUpdateStatus(uid string, st status.Status) {
	u.locker.Lock()
	defer u.locker.Unlock()

	if u.uidmap[uid] == nil {
		panic(user.ErrUserNotExists)
	}
	ok, err := status.NormalOrBannedService.IsAvailable(st)
	if err != nil {
		panic(err)
	}
	u.uidmap[uid].Banned = !ok
	u.mustSave()

}
func (u *Users) MustCreateStatus(uid string) {
	u.locker.Lock()
	defer u.locker.Unlock()
	if u.uidmap[uid] != nil {
		panic(user.ErrUserExists)
	}
	newuser := NewUser()
	newuser.UID = uid
	u.addUser(newuser)
	u.mustSave()
}
func (u *Users) MustRemoveStatus(uid string) {
	u.locker.Lock()
	defer u.locker.Unlock()
	if u.uidmap[uid] == nil {
		panic(user.ErrUserNotExists)
	}
	u.removeUser(uid)
	u.mustSave()

}
func (u *Users) getAfterLast(last string, users []string, reverse bool) []string {
	if reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(users)))
	} else {
		sort.Strings(users)
	}
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

func (u *Users) MustListUsersByStatus(last string, limit int, reverse bool, statuses ...status.Status) []string {
	u.locker.RLock()
	defer u.locker.RUnlock()
	m := map[bool]bool{}
	for k := range statuses {
		ok, err := u.IsAvailable(statuses[k])
		if err != nil {
			return nil
		}
		m[!ok] = true
	}
	users := []string{}
	for k := range u.uidmap {
		if m[u.uidmap[k].Banned] {
			users = append(users, k)
		}
	}
	result := u.getAfterLast(last, users, reverse)
	if limit > 0 && limit < len(result) {
		return result[:limit]
	}
	return result
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
func (u *Users) MustRoles(uid string) *role.Roles {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		panic(user.ErrUserNotExists)
	}
	return us.Roles.Clone()
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
func (u *Users) Accounts(uid string) (*user.Accounts, error) {
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

func (u *Users) MustCurrentTerm(uid string) string {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		panic(user.ErrUserNotExists)
	}
	return us.Term

}
func (u *Users) MustStartNewTerm(uid string) string {
	u.locker.Lock()
	defer u.locker.Unlock()
	us := u.uidmap[uid]
	if us == nil {
		panic(user.ErrUserNotExists)
	}
	term, err := u.idFactory()
	if err != nil {
		panic(err)
	}
	us.Term = term
	return term
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
