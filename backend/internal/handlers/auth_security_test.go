package handlers

import (
 "database/sql/driver"
 "encoding/json"
 "errors"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/DATA-DOG/go-sqlmock"
 "github.com/gofiber/fiber/v2"
 "github.com/jmoiron/sqlx"
 "sikapj-backend/internal/auth"
 "sikapj-backend/internal/config"
)
func TestPasswordChangeRevokesSessionsAtomically(t *testing.T){
 db,m:=testDB(t);hash,_:=auth.HashPassword("old-password");m.ExpectQuery("SELECT password_hash").WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(hash));m.ExpectBegin();m.ExpectExec("UPDATE sikapj.users SET .*session_version=session_version\\+1").WillReturnResult(sqlmock.NewResult(0,1));m.ExpectExec("INSERT INTO sikapj.audit_log").WithArgs(1,"CHANGE_PASSWORD",jsonArg{},sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1,1));m.ExpectCommit();h:=AuthHandler{DB:db};r:=httptest.NewRequest("PUT","/password",strings.NewReader(`{"current_password":"old-password","new_password":"new-password"}`));r.Header.Set("Content-Type","application/json");res,e:=testApp(h.ChangePassword).Test(r);if e!=nil{t.Fatal(e)};defer res.Body.Close();if res.StatusCode!=200{t.Errorf("status %d",res.StatusCode)};if e=m.ExpectationsWereMet();e!=nil{t.Error(e)}
}
type jsonArg struct{}
func(jsonArg) Match(v driver.Value) bool { s,ok:=v.(string); return ok && json.Valid([]byte(s)) }
func testDB(t *testing.T)(*sqlx.DB,sqlmock.Sqlmock){ t.Helper(); d,m,e:=sqlmock.New();if e!=nil{t.Fatal(e)};t.Cleanup(func(){d.Close()});return sqlx.NewDb(d,"sqlmock"),m }
func testApp(h fiber.Handler)*fiber.App{ a:=fiber.New();a.Use(func(c *fiber.Ctx)error{c.Locals("user_id",1);c.Locals("role","ADMIN");c.Locals("nip","199003202014022005");return c.Next()});a.All("/*",h);return a }
func TestLoginAuditIsJSONAndAtomic(t *testing.T){
 for _,fail:=range []bool{false,true}{t.Run(map[bool]string{false:"success",true:"audit_failure"}[fail],func(t *testing.T){
 db,m:=testDB(t);hash,_:=auth.HashPassword("valid-password");m.ExpectQuery("SELECT .*FROM sikapj.users").WillReturnRows(sqlmock.NewRows([]string{"id","username","password_hash","role","is_active"}).AddRow(1,"admin",hash,"ADMIN",true))
 m.ExpectBegin();m.ExpectExec("UPDATE sikapj.users SET last_login").WillReturnResult(sqlmock.NewResult(0,1));ex:=m.ExpectExec("INSERT INTO sikapj.audit_log").WithArgs(1,"LOGIN",jsonArg{},sqlmock.AnyArg());if fail{ex.WillReturnError(errors.New("audit unavailable"));m.ExpectRollback()}else{ex.WillReturnResult(sqlmock.NewResult(1,1));m.ExpectCommit()}
 h:=AuthHandler{DB:db,Cfg:&config.Config{JWTSecret:strings.Repeat("x",32),JWTExpirationHours:1}};r:=httptest.NewRequest("POST","/login",strings.NewReader(`{"username":"admin","password":"valid-password"}`));r.Header.Set("Content-Type","application/json");res,e:=testApp(h.Login).Test(r);if e!=nil{t.Fatal(e)};defer res.Body.Close();want:=200;if fail{want=500};if res.StatusCode!=want{t.Errorf("status %d want %d",res.StatusCode,want)};if e=m.ExpectationsWereMet();e!=nil{t.Error(e)}
 })}
}
