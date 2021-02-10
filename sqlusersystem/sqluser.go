package sqlusersystem

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"crypto/rand"
	"crypto/sha256"

	"github.com/herb-go/datasource/sql/db"
	"github.com/herb-go/datasource/sql/querybuilder"
	"github.com/herb-go/datasource/sql/querybuilder/modelmapper"
	"github.com/herb-go/user"
	"github.com/herb-go/user/status"
)

//RandomBytesLength bytes length for RandomBytes function.
var RandomBytesLength = 32

//ErrHashMethodNotFound error raised when password hash method not found.
var ErrHashMethodNotFound = errors.New("password hash method not found")

//HashFunc interaface of pasword hash func
type HashFunc func(key string, salt string, password string) ([]byte, error)

//DefaultAccountMapperName default database table name for module account.
var DefaultAccountMapperName = "account"

//DefaultPasswordMapperName default database table name for module password.
var DefaultPasswordMapperName = "password"

//DefaultTokenMapperName default database table name for module token.
var DefaultTokenMapperName = "token"

//DefaultUserMapperName default database table name for module user.
var DefaultUserMapperName = "user"

//DefaultHashMethod default hash method when created password data.
var DefaultHashMethod = "sha256"

//HashFuncMap all available password hash func.
//You can insert custom hash func into this map.
var HashFuncMap = map[string]HashFunc{
	"sha256": func(key string, salt string, password string) ([]byte, error) {
		var val = []byte(key + salt + password)
		var s256 = sha256.New()
		s256.Write(val)
		val = s256.Sum(nil)
		s256.Write(val)
		return []byte(hex.EncodeToString(s256.Sum(nil))), nil
	},
}

//New create User framework
func New() *User {
	return &User{
		Tables: Tables{
			AccountMapperName:  DefaultAccountMapperName,
			PasswordMapperName: DefaultPasswordMapperName,
			TokenMapperName:    DefaultTokenMapperName,
			UserMapperName:     DefaultUserMapperName,
		},
		HashMethod:     DefaultHashMethod,
		TokenGenerater: Timestamp,
		SaltGenerater:  RandomBytes,
	}
}

//Tables struct stores table info.
type Tables struct {
	AccountMapperName  string
	PasswordMapperName string
	TokenMapperName    string
	UserMapperName     string
}

//RandomBytes string generater return random bytes.
//Default length is 32 byte.You can change default length by change sqluesr.RandomBytesLength .
func RandomBytes() (string, error) {
	var bytes = make([]byte, RandomBytesLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

//Timestamp string generater return timestamp in nano.
func Timestamp() (string, error) {
	return strconv.FormatInt(time.Now().UnixNano(), 10), nil
}

//User main struct of sqluser module.
type User struct {
	//DB database used.
	DB db.Database
	//Tables table name info.
	Tables Tables
	//UIDGenerater string generater for uid
	//default value is uuid
	UIDGenerater func() (string, error)
	//TokenGenerater string generater for usertoken
	//default value is timestamp
	TokenGenerater func() (string, error)
	//SaltGenerater string generater for salt
	//default value is 32 byte length random bytes.
	SaltGenerater func() (string, error)
	//HashMethod hash method which used to generate new salt.
	//default value is sha256
	HashMethod string
	//PasswordKey static key used in password hash generater.
	//default value is empty.
	//You can change this value after sqluser init.
	PasswordKey string
	//QueryBuilder sql query builder
	QueryBuilder *querybuilder.Builder
}

//AddTablePrefix add prefix to user table names.
func (u *User) AddTablePrefix(prefix string) {
	u.Tables.AccountMapperName = prefix + u.Tables.AccountMapperName
	u.Tables.PasswordMapperName = prefix + u.Tables.PasswordMapperName
	u.Tables.TokenMapperName = prefix + u.Tables.TokenMapperName
	u.Tables.UserMapperName = prefix + u.Tables.UserMapperName
}

//AccountTableName return actual account database table name.
func (u *User) AccountTableName() string {
	return u.DB.BuildTableName(u.Tables.AccountMapperName)
}

//PasswordTableName return actual password database table name.
func (u *User) PasswordTableName() string {
	return u.DB.BuildTableName(u.Tables.PasswordMapperName)
}

//TokenTableName return actual token database table name.
func (u *User) TokenTableName() string {
	return u.DB.BuildTableName(u.Tables.TokenMapperName)
}

//UserTableName return actual user database table name.
func (u *User) UserTableName() string {
	return u.DB.BuildTableName(u.Tables.UserMapperName)
}

//Account return account mapper
func (u *User) Account() *AccountMapper {
	return &AccountMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.AccountMapperName)),
		User:        u,
	}
}

//Password return password mapper
func (u *User) Password() *PasswordMapper {
	return &PasswordMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.PasswordMapperName)),
		User:        u,
	}
}

//Token return token mapper
func (u *User) Token() *TokenMapper {
	return &TokenMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.TokenMapperName)),
		User:        u,
	}
}

//User return user mapper
func (u *User) User() *UserMapper {
	return &UserMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.UserMapperName)),
		User:        u,
	}
}

//AccountMapper account mapper
type AccountMapper struct {
	*modelmapper.ModelMapper
	User *User
}

//Accounts return accounts of give uid.
func (a *AccountMapper) Accounts(uid string) (*user.Accounts, error) {
	query := a.User.QueryBuilder
	var result = []*user.Account{}
	Select := query.NewSelectQuery()
	Select.Select.Add("account.keyword", "account.account")
	Select.From.AddAlias("account", a.TableName())
	Select.Where.Condition = query.Equal("account.uid", uid)
	rows, err := Select.QueryRows(a.DB())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := user.NewAccount()
		err := Select.Result().
			Bind("account.keyword", &v.Keyword).
			Bind("account.account", &v.Account).
			ScanFrom(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	accounts := user.Accounts(result)
	return &accounts, nil
}

//AccountToUID query uid by user account.
//Return user id and any error if raised.
//Return empty string as userid if account not found.
func (a *AccountMapper) AccountToUID(account *user.Account) (uid string, err error) {
	model, err := a.Find(account.Keyword, account.Account)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return model.UID, nil
}

func (a *AccountMapper) Start() error {
	return nil
}

//Stop stop service
func (a *AccountMapper) Stop() error {
	return nil
}

//Purge purge user data cache
func (a *AccountMapper) Purge(string) error {
	return nil
}

//Unbind unbind account from user.
//Return any error if raised.
func (a *AccountMapper) Unbind(uid string, account *user.Account) error {
	query := a.User.QueryBuilder
	Delete := query.NewDeleteQuery(a.TableName())
	Delete.Where.Condition = query.And(
		query.Equal("account.uid", uid),
		query.Equal("account.keyword", account.Keyword),
		query.Equal("account.account", account.Account),
	)
	r, err := Delete.Query().Exec(a.DB())
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return user.ErrAccountUnbindingNotExists
	}
	return nil
}

//Bind bind account to user.
//Return any error if raised.
//If account exists, error user.ErrAccountBindingExists will raised.
func (a *AccountMapper) Bind(uid string, account *user.Account) error {
	query := a.User.QueryBuilder
	tx, err := a.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var u = ""
	Select := query.NewSelectQuery()
	Select.Select.Add("account.uid")
	Select.From.AddAlias("account", a.TableName())
	Select.Where.Condition = query.And(
		query.Equal("keyword", account.Keyword),
		query.Equal("account", account.Account),
	)
	row := Select.QueryRow(a.DB())
	err = row.Scan(&u)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		return user.ErrAccountBindingExists

	}

	var CreatedTime = time.Now().Unix()
	Insert := query.NewInsertQuery(a.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("keyword", account.Keyword).
		Add("account", account.Account).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//Insert create new user with given account.

//Find find account by given keyword and account.
//Return account model and any error if raised.
func (a *AccountMapper) Find(keyword string, account string) (*AccountModel, error) {
	query := a.User.QueryBuilder
	var result = &AccountModel{}
	if account == "" {
		return nil, sql.ErrNoRows
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("uid", "keyword", "account", "created_time")
	Select.From.Add(a.TableName())
	Select.Where.Condition = query.And(
		query.Equal("keyword", keyword),
		query.Equal("account", account),
	)
	row := Select.QueryRow(a.DB())
	err := Select.Result().
		Bind("uid", &result.UID).
		Bind("keyword", &result.Keyword).
		Bind("account", &result.Account).
		Bind("created_time", &result.CreatedTime).
		ScanFrom(row)
	return result, err
}

//BindAccount bind account to user.
//Return any error if rasied.
//If account exists, error user.ErrAccountBindingExists will raised.
func (a *AccountMapper) BindAccount(uid string, account *user.Account) error {
	return a.Bind(uid, account)
}

//UnbindAccount unbind account from user.
//Return any error if rasied.
func (a *AccountMapper) UnbindAccount(uid string, account *user.Account) error {
	return a.Unbind(uid, account)
}

//AccountModel account data model
type AccountModel struct {
	//UID user id.
	UID string
	//Keyword account keyword.
	Keyword string
	//Account account name.
	Account string
	//CreatedTime created timestamp in second.
	CreatedTime int64
}

//PasswordMapper password mapper
type PasswordMapper struct {
	*modelmapper.ModelMapper
	User *User
}

//Start start service
func (p *PasswordMapper) Start() error {
	return nil
}

//Stop stop service
func (p *PasswordMapper) Stop() error {
	return nil
}

//Purge purge user data cache
func (p *PasswordMapper) Purge(string) error {
	return nil
}

//PasswordChangeable return password changeable
func (p *PasswordMapper) PasswordChangeable() bool {
	return true
}

//Find find password model by userd id.
//Return any error if raised.
func (p *PasswordMapper) Find(uid string) (PasswordModel, error) {
	query := p.User.QueryBuilder
	var result = PasswordModel{}
	if uid == "" {
		return result, sql.ErrNoRows
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("password.hash_method", "password.salt", "password.password", "password.updated_time")
	Select.From.AddAlias("password", p.TableName())
	Select.Where.Condition = query.Equal("uid", uid)
	q := Select.Query()
	row := p.DB().QueryRow(q.QueryCommand(), q.QueryArgs()...)
	result.UID = uid
	args := Select.Result().
		Bind("password.hash_method", &result.HashMethod).
		Bind("password.salt", &result.Salt).
		Bind("password.password", &result.Password).
		Bind("password.updated_time", &result.UpdatedTime).
		Pointers()

	err := row.Scan(args...)
	return result, err
}

//InsertOrUpdate insert or update password model.
//Return any error if raised.
func (p *PasswordMapper) InsertOrUpdate(model *PasswordModel) error {
	query := p.User.QueryBuilder

	tx, err := p.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	Update := query.NewUpdateQuery(p.TableName())
	Update.Update.
		Add("hash_method", model.HashMethod).
		Add("salt", model.Salt).
		Add("password", model.Password).
		Add("updated_time", model.UpdatedTime)
	Update.Where.Condition = query.Equal("uid", model.UID)
	r, err := Update.Query().Exec(tx)

	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	Insert := query.NewInsertQuery(p.TableName())
	Insert.Insert.
		Add("uid", model.UID).
		Add("hash_method", model.HashMethod).
		Add("salt", model.Salt).
		Add("password", model.Password).
		Add("updated_time", model.UpdatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//VerifyPassword Verify user password.
//Return verify and any error if raised.
//if user not found,error user.ErrUserNotExists will be raised.
func (p *PasswordMapper) VerifyPassword(uid string, password string) (bool, error) {
	model, err := p.Find(uid)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	hash := HashFuncMap[model.HashMethod]
	if hash == nil {
		return false, ErrHashMethodNotFound
	}
	hashed, err := hash(p.User.PasswordKey, model.Salt, password)
	if err != nil {
		return false, err
	}
	return bytes.Compare(hashed, model.Password) == 0, nil
}

//UpdatePassword update user password.If user password does not exist,new password record will be created.
//Return any error if raised.
func (p *PasswordMapper) UpdatePassword(uid string, password string) error {
	salt, err := p.User.SaltGenerater()
	if err != nil {
		return err
	}
	hash := HashFuncMap[p.User.HashMethod]
	if hash == nil {
		return ErrHashMethodNotFound
	}
	hashed, err := hash(p.User.PasswordKey, salt, password)
	if err != nil {
		return err
	}
	model := &PasswordModel{
		UID:         uid,
		HashMethod:  p.User.HashMethod,
		Salt:        salt,
		Password:    hashed,
		UpdatedTime: time.Now().Unix(),
	}
	return p.InsertOrUpdate(model)
}

//PasswordModel password data model
type PasswordModel struct {
	//UID user id.
	UID string
	//HashMethod hash method to verify this password.
	HashMethod string
	//Salt random salt.
	Salt string
	//Password hashed password data.
	Password []byte
	//UpdatedTime updated timestamp in second.
	UpdatedTime int64
}

//TokenMapper token mapper
type TokenMapper struct {
	*modelmapper.ModelMapper
	User *User
}

func (t *TokenMapper) CurrentTerm(uid string) (string, error) {
	query := t.User.QueryBuilder
	Select := query.NewSelectQuery()
	Select.Select.Add("token.token")
	Select.From.AddAlias("token", t.TableName())
	Select.Where.Condition = query.Equal("token.uid", uid)
	row := Select.QueryRow(t.DB())
	var token string
	err := row.Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return token, nil
}
func (t *TokenMapper) StartNewTerm(uid string) (string, error) {
	token, err := t.User.TokenGenerater()
	if err != nil {
		return "", err
	}
	return token, t.InsertOrUpdate(uid, token)
}

//Start start service
func (t *TokenMapper) Start() error {
	return nil
}

//Stop stop service
func (t *TokenMapper) Stop() error {
	return nil
}

//Purge purge user data cache
func (t *TokenMapper) Purge(string) error {
	return nil
}

//InsertOrUpdate insert or update user token record.
func (t *TokenMapper) InsertOrUpdate(uid string, token string) error {
	query := t.User.QueryBuilder

	tx, err := t.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var CreatedTime = time.Now().Unix()
	Update := query.NewUpdateQuery(t.TableName())
	Update.Update.
		Add("token", token).
		Add("updated_time", CreatedTime)
	Update.Where.Condition = query.Equal("uid", uid)
	r, err := Update.Query().Exec(tx)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	Insert := query.NewInsertQuery(t.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("token", token).
		Add("updated_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//TokenModel token data model
type TokenModel struct {
	//UID user id
	UID string
	//Token current user token
	Token string
	//UpdatedTime updated timestamp in second.
	UpdatedTime string
}

//UserMapper user mapper
type UserMapper struct {
	*modelmapper.ModelMapper
	User *User
}

//IsAvailable check is status available
func (u *UserMapper) IsAvailable(userstats status.Status) (bool, error) {
	return status.NormalOrBannedService.IsAvailable(userstats)
}

//Label get status label
//Empty string will be returned if status invalid
func (u *UserMapper) Label(userstats status.Status) (string, error) {
	return status.NormalOrBannedService.Label(userstats)
}

func (u *UserMapper) LoadStatus(uid string) (status.Status, error) {
	var userstatus status.Status
	query := u.User.QueryBuilder
	Select := query.NewSelectQuery()
	Select.Select.Add("user.status")
	Select.From.AddAlias("user", u.TableName())
	Select.Where.Condition = query.Equal("user.uid", uid)
	row := Select.QueryRow(u.DB())
	err := row.Scan(&userstatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, user.ErrUserNotExists
		}
		return 0, err
	}
	return userstatus, nil
}
func (u *UserMapper) UpdateStatus(uid string, userstatus status.Status) error {
	query := u.User.QueryBuilder
	tx, err := u.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var CreatedTime = time.Now().Unix()
	Update := query.NewUpdateQuery(u.TableName())
	Update.Update.
		Add("status", userstatus).
		Add("updated_time", CreatedTime)
	Update.Where.Condition = query.Equal("uid", uid)
	r, err := Update.Query().Exec(tx)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	return user.ErrUserNotExists
}
func (u *UserMapper) CreateStatus(uid string) error {
	query := u.User.QueryBuilder
	tx, err := u.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var CreatedTime = time.Now().Unix()
	Insert := query.NewInsertQuery(u.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("status", status.StatusUnkown).
		Add("updated_time", CreatedTime).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		if query.IsDuplicate(err) {
			return user.ErrUserExists
		}
		return err
	}
	return tx.Commit()
}
func (u *UserMapper) RemoveStatus(uid string) error {
	query := u.User.QueryBuilder
	Delete := query.NewDeleteQuery(u.TableName())
	Delete.Where.Condition = query.Equal("uid", uid)
	result, err := Delete.Query().Exec(u.DB())
	if err != nil {
		return err
	}
	a, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if a == 0 {
		return user.ErrUserNotExists
	}
	return nil
}
func (u *UserMapper) ListUsersByStatus(last string, limit int, reverse bool, statuses ...status.Status) ([]string, error) {
	query := u.User.QueryBuilder
	Select := query.NewSelectQuery()
	Select.Select.Add("user.uid")
	Select.From.AddAlias("user", u.TableName())
	if last != "" {
		if reverse {
			Select.Where.Condition = query.New("user.uid < ?", last)
		} else {
			Select.Where.Condition = query.New("user.uid > ?", last)
		}
	}
	if len(statuses) > 0 {
		Select.Where.Condition = Select.Where.Condition.And(query.In("user.status", statuses))
	}
	if limit != 0 {
		Select.Limit.Limit = &limit
	}
	Select.OrderBy.Add("user.uid", !reverse)
	rows, err := Select.QueryRows(u.DB())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var uid string
		err = rows.Scan(&uid)
		if err != nil {
			return nil, err
		}
		result = append(result, uid)
	}
	return result, nil
}
func (u *UserMapper) Purge(uid string) error {
	return nil
}
func (u *UserMapper) Start() error {
	return nil
}
func (u *UserMapper) Stop() error {
	return nil
}

//UserModel user data model
type UserModel struct {
	//UID user id
	UID string
	//CreatedTime created timestamp in second
	CreatedTime int64
	//UpdateTIme updated timestamp in second
	UpdateTIme int64
	//Status user status
	Status int
}
