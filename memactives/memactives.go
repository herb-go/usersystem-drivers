package memactives

import (
	"sync"
	"time"

	"github.com/herb-go/uniqueid"

	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/activesessions"
)

type UserData map[string]*activesessions.Active

func (d *UserData) OnSessionActive(session *usersystem.Session) {
	sn := session.Payloads.LoadString(activesessions.PayloadSerialNumber)
	if sn == "" {
		return
	}
	active, ok := (*d)[sn]
	if !ok {
		active = &activesessions.Active{}
		active.SessionID = session.ID
		(*d)[sn] = active
	}
	active.LastActive = time.Now()
}

func (d *UserData) PurgeActiveSession(serialnumber string) {
	delete(*d, serialnumber)
}
func (d *UserData) GetActiveSessions(from time.Time) map[string]*activesessions.Active {
	result := map[string]*activesessions.Active{}
	for k, v := range *d {
		if v.LastActive.After(from) {
			result[k] = v
		}
	}
	return result
}
func NewUserdata() *UserData {
	return &UserData{}
}

type Store struct {
	CreatedTime time.Time
	Data        map[string]*UserData
}

func (s *Store) OnSessionActive(session *usersystem.Session) {
	uid := session.UID()
	ud, ok := s.Data[uid]
	if !ok {
		ud = NewUserdata()
		s.Data[uid] = ud
	}
	ud.OnSessionActive(session)
}
func (s *Store) PurgeActiveSession(uid string, serialnumber string) {
	ud, ok := s.Data[uid]
	if !ok {
		return
	}
	ud.PurgeActiveSession(serialnumber)

}
func (s *Store) GetActiveSessions(uid string, from time.Time) map[string]*activesessions.Active {
	ud, ok := s.Data[uid]
	if !ok {
		return nil
	}
	return ud.GetActiveSessions(from)
}
func NewStore() *Store {
	return &Store{
		Data: map[string]*UserData{},
	}
}

type StoreList struct {
	Ticker   *time.Ticker
	Duration time.Duration
	Locker   sync.RWMutex
	List     []*Store
}

func (l *StoreList) OnSessionActive(session *usersystem.Session) {
	if session.ID == "" || session.UID() == "" || session.Payloads.LoadString(activesessions.PayloadSerialNumber) == "" {
		return
	}
	l.Locker.Lock()
	defer l.Locker.Unlock()
	l.List[0].OnSessionActive(session)
}
func (l *StoreList) GetActiveSessions(uid string) []*activesessions.Active {
	l.Locker.RLock()
	defer l.Locker.RUnlock()
	from := time.Now().Add(-l.Duration)
	var all = map[string]*activesessions.Active{}
	for i := len(l.List) - 1; i >= 0; i-- {
		data := l.List[i].GetActiveSessions(uid, from)
		for k := range data {
			if _, ok := all[k]; !ok {
				all[k] = data[k]
			}
		}
	}
	result := make([]*activesessions.Active, 0, len(all))
	for k := range all {
		result = append(result, all[k])
	}
	return result
}
func (l *StoreList) PurgeActiveSession(uid string, serialnumber string) {
	l.Locker.Lock()
	defer l.Locker.Unlock()
	for _, v := range l.List {
		v.PurgeActiveSession(uid, serialnumber)
	}
}
func (l *StoreList) Start() error {
	l.Ticker = time.NewTicker(l.Duration)
	go func() {
		for _ = range l.Ticker.C {
			l.Update()
		}
	}()
	return nil
}
func (l *StoreList) Stop() error {
	l.Ticker.Stop()
	return nil
}
func (l *StoreList) Update() {
	l.Locker.Lock()
	defer l.Locker.Unlock()
	list := make([]*Store, 0, len(l.List))
	list = append(list, NewStore())
	list[0].CreatedTime = time.Now()
	deadline := list[0].CreatedTime.Add(-l.Duration)
	for k := range l.List {
		if l.List[k].CreatedTime.After(deadline) {
			list = append(list, l.List[k])
		}
	}
	l.List = list
}
func NewStoreList() *StoreList {
	return &StoreList{}
}

type Service struct {
	Stores map[usersystem.SessionType]*StoreList
}

func (s *Service) MustConfig(st usersystem.SessionType) *activesessions.Config {
	stores, ok := s.Stores[st]
	if !ok {
		return &activesessions.Config{
			Supported: false,
		}
	}
	return &activesessions.Config{
		Supported: true,
		Duration:  stores.Duration,
	}
}
func (s *Service) MustOnSessionActive(session *usersystem.Session) {
	if session == nil {
		return
	}
	stores, ok := s.Stores[session.Type]
	if !ok {
		return
	}
	stores.OnSessionActive(session)
	return
}
func (s *Service) MustGetActiveSessions(st usersystem.SessionType, uid string) []*activesessions.Active {
	stores, ok := s.Stores[st]
	if !ok {
		return nil
	}
	return stores.GetActiveSessions(uid)
}
func (s *Service) MustPurgeActiveSession(st usersystem.SessionType, uid string, serialnumber string) {
	if uid == "" || serialnumber == "" {
		return
	}
	stores, ok := s.Stores[st]
	if !ok {
		return
	}
	stores.PurgeActiveSession(uid, serialnumber)
	return
}
func (s *Service) MustCreateSerialNumber() string {
	return uniqueid.DefaultGenerator.MustGenerateID()
}
func (s *Service) Start() error {
	var err error
	for _, v := range s.Stores {
		err = v.Start()
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) Stop() error {
	for _, v := range s.Stores {
		v.Stop()
	}
	return nil
}
func (s *Service) Execute(us *usersystem.UserSystem) error {
	as := activesessions.MustGetModule(us)
	if as == nil {
		return nil
	}
	as.Service = s
	return nil
}
func NewService() *Service {
	return &Service{
		Stores: map[usersystem.SessionType]*StoreList{},
	}
}
