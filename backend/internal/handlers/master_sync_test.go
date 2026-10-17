package handlers

import (
 "encoding/json"
 "errors"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/DATA-DOG/go-sqlmock"
 "github.com/gofiber/fiber/v2"
)

func TestMasterCommitRemainsSuccessfulWhenAlertSyncFails(t *testing.T) {
 for _,action:=range []string{"create","update","toggle"} {t.Run(action,func(t *testing.T){
  db,m:=testDB(t);h:=&PegawaiHandler{DB:db};var handler fiber.Handler
  m.ExpectBegin()
  switch action {
  case "create": handler=h.CreatePegawai;m.ExpectQuery("INSERT INTO sikapj.users .*must_change_password.*TRUE, TRUE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2));m.ExpectExec("INSERT INTO sikapj.pegawai").WillReturnResult(sqlmock.NewResult(0,1));m.ExpectExec("INSERT INTO sikapj.audit_log").WithArgs(1,"CREATE_PEGAWAI",jsonArg{},sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1,1))
  case "update":handler=h.UpdatePegawai;m.ExpectQuery("UPDATE sikapj.pegawai").WillReturnRows(sqlmock.NewRows([]string{"is_active"}).AddRow(true));m.ExpectExec("UPDATE sikapj.users").WillReturnResult(sqlmock.NewResult(0,1));m.ExpectExec("INSERT INTO sikapj.audit_log").WithArgs(1,"UPDATE_PEGAWAI",jsonArg{},sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1,1))
  case "toggle":handler=h.ToggleStatusPegawai;m.ExpectQuery("UPDATE sikapj.pegawai").WillReturnRows(sqlmock.NewRows([]string{"is_active"}).AddRow(false));m.ExpectExec("UPDATE sikapj.users").WillReturnResult(sqlmock.NewResult(0,1));m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnResult(sqlmock.NewResult(1,1))
  }
  m.ExpectCommit()
  m.ExpectQuery("SELECT value FROM sikapj.settings").WillReturnError(errors.New("sync unavailable"))
  // Sync implementations may fail fast or continue with the configured threshold.
  m.ExpectQuery("SELECT .*FROM sikapj.pegawai").WillReturnError(errors.New("sync unavailable"))
  b,_:=json.Marshal(validEmployeePayload());r:=httptest.NewRequest("PUT","/199003202014022005",strings.NewReader(string(b)));r.Header.Set("Content-Type","application/json");res,e:=testApp(handler).Test(r);if e!=nil{t.Fatal(e)};defer res.Body.Close();if res.StatusCode!=200{t.Fatalf("committed mutation returned %d",res.StatusCode)}
 })}
}
