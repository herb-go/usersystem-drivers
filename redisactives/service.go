package redisactives

import (
	"time"

	"github.com/herb-go/herbdata/datautil"
	"github.com/vmihailenco/msgpack"

	"github.com/gomodule/redigo/redis"
	"github.com/herb-go/uniqueid"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/activesessions"
)

type Service struct {
	configs  map[usersystem.SessionType]*activesessions.Config
	Pool     *redis.Pool
	Prefix   string
	Interval int64
}

func (s *Service) MustConfig(st usersystem.SessionType) *activesessions.Config {
	if s.configs[st] == nil {
		return &activesessions.Config{
			Supported: false,
		}
	}
	return s.configs[st]
}
func (s *Service) MustOnSessionActive(session *usersystem.Session) {
	if session == nil {
		return
	}
	config, ok := s.configs[session.Type]
	if !ok {
		return
	}
	sn := session.Payloads.LoadString(activesessions.PayloadSerialNumber)
	if sn == "" {
		return
	}
	duration := config.Duration
	uid := session.UID()
	now := time.Now().Unix()
	key := s.sessionkey(session.Type, uid, now, duration, false)
	sessionid := session.ID
	err := s.activeEntry(key, sn, sessionid, now, duration)
	if err != nil {
		panic(err)
	}
}
func (s *Service) sessionkey(st usersystem.SessionType, uid string, now int64, duration time.Duration, prev bool) []byte {
	return datautil.Join(nil, []byte(s.Prefix), []byte(st), []byte(uid), timekey(now, s.configs[st].Duration, prev))
}
func (s *Service) getData(conn redis.Conn, key []byte, sn string, sessionid string) (*entry, error) {
	var userentry = &entry{}
	data, err := redis.Bytes(conn.Do("HGET", key, sn))
	if err != redis.ErrNil && data != nil {
		err = msgpack.Unmarshal(data, userentry)
		if err != nil {
			return nil, err
		}
	} else {
		userentry.ID = sessionid
	}
	return userentry, nil
}
func (s *Service) updateEntry(conn redis.Conn, key []byte, sn string, duration time.Duration, userentry *entry) error {
	data, err := msgpack.Marshal(userentry)
	if err != nil {
		return err
	}
	_, err = conn.Do("Multi")
	if err != nil {
		return err
	}
	err = conn.Send("HSET", key, sn, data)
	if err != nil {
		return err
	}
	err = conn.Send("EXPIRE", key, int64(duration/time.Second))
	if err != nil {
		return err
	}
	_, err = conn.Do("Exec")
	return err
}
func (s *Service) activeEntry(key []byte, sn string, sessionid string, now int64, duration time.Duration) error {
	var userentry = &entry{}
	conn := s.Pool.Get()
	defer conn.Close()
	userentry, err := s.getData(conn, key, sn, sessionid)
	if err != nil {
		return err
	}
	if userentry.LastActive+s.Interval < now {
		userentry.LastActive = now
		err = s.updateEntry(conn, key, sn, duration, userentry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) unmarshalEntries(data [][]byte, err error) ([]*entry, error) {
	if err != nil {
		return nil, err
	}
	var entrylist = []*entry{}
	for _, v := range data {
		if v != nil {
			entry := &entry{}
			err := msgpack.Unmarshal(v, entry)
			if err != nil {
				return nil, err
			}
			entrylist = append(entrylist, entry)
		}
	}
	return entrylist, nil
}
func (s *Service) MustGetActiveSessions(st usersystem.SessionType, uid string) []*activesessions.Active {
	if uid == "" || s.configs[st] == nil {
		return nil
	}
	conn := s.Pool.Get()
	defer conn.Close()
	var loaded = map[string]bool{}
	now := time.Now().Unix()
	duration := s.configs[st].Duration
	key := s.sessionkey(st, uid, now, duration, false)
	keyprev := s.sessionkey(st, uid, now, duration, true)
	resultnow, err := s.unmarshalEntries(redis.ByteSlices(conn.Do("HVALS", key)))
	if err != nil {
		panic(err)
	}
	resultprev, err := s.unmarshalEntries(redis.ByteSlices(conn.Do("HVALS", keyprev)))
	if err != nil {
		panic(err)
	}
	var result = []*activesessions.Active{}
	for _, v := range resultnow {
		if loaded[v.ID] == false && v.LastActive < now+int64(duration/time.Second) {
			loaded[v.ID] = true
			result = append(result, v.Convert())
		}
	}
	for _, v := range resultprev {
		if loaded[v.ID] == false && v.LastActive+s.Interval < now+int64(duration/time.Second) {
			loaded[v.ID] = true
			result = append(result, v.Convert())
		}
	}
	return result
}

func (s *Service) MustPurgeActiveSession(st usersystem.SessionType, uid string, serialnumber string) {
	if uid == "" || s.configs[st] == nil || serialnumber == "" {
		return
	}
	conn := s.Pool.Get()
	defer conn.Close()
	now := time.Now().Unix()
	duration := s.configs[st].Duration
	key := s.sessionkey(st, uid, now, duration, false)
	keyprev := s.sessionkey(st, uid, now, duration, true)
	_, err := conn.Do("HDEL", key, serialnumber)
	if err != nil {
		panic(err)
	}
	_, err = conn.Do("HDEL", keyprev, serialnumber)
	if err != nil {
		panic(err)
	}
}
func (s *Service) MustCreateSerialNumber() string {
	return uniqueid.DefaultGenerator.MustGenerateID()
}
func (s *Service) Start() error {
	return nil
}
func (s *Service) Stop() error {
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
		configs: map[usersystem.SessionType]*activesessions.Config{},
	}
}
