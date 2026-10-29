package main

import (
 "errors"
 "strings"
 "testing"
)

type psRunner struct{ calls [][]string; fail bool }
func (f *psRunner) Run(name string,args ...string)(string,error){
 f.calls=append(f.calls,append([]string{name},args...))
 s:=strings.Join(append([]string{name},args...)," ")
 if strings.Contains(s," ps --format "){
  if f.fail{return "broken",errors.New("docker unavailable")}
  return "{\"ID\":\"abc123\",\"Names\":\"z-ollama\",\"Image\":\"ollama:latest\",\"Command\":\"serve\",\"Ports\":\"127.0.0.1:28134->11434/tcp\"}\n{\"ID\":\"def456\",\"Names\":\"a-dashboard\",\"Image\":\"dash:v1\",\"Command\":\"uvicorn\",\"Ports\":\"127.0.0.1:28080->8080/tcp\"}",nil
 }
 if strings.Contains(s," inspect "){return "sha256:1234567890abcdef",nil}
 return "",nil
}
func TestContainerURL(t *testing.T){
 cases:=[]struct{ports,want string}{
 {"127.0.0.1:28080->8080/tcp","http://127.0.0.1:28080"},
 {"0.0.0.0:8443->443/tcp","https://localhost:8443"},
 {"[::]:28188->8188/tcp","http://localhost:28188"},
 }
 for _,tt:=range cases{
  got,err:=containerURL(tt.ports);if err!=nil||got!=tt.want{t.Fatalf("%q => %q, %v; want %q",tt.ports,got,err,tt.want)}
 }
 if _,err:=containerURL("8188/tcp");err==nil{t.Fatal("expected no published port error")}
}
func TestPSOrderAndInspect(t *testing.T){
 f:=&psRunner{}
 a:=NewApp(f)
 if err:=a.Run([]string{"ps"});err!=nil{t.Fatal(err)}
 if len(f.calls)!=3{t.Fatalf("expected ps plus 2 inspect calls, got %d",len(f.calls))}
 if !strings.Contains(strings.Join(f.calls[1]," "),"def456"){t.Fatal("containers not sorted by name")}
}
func TestOpenSelection(t *testing.T){
 for _,sel:=range []string{"1","a-dashboard"}{
  f:=&psRunner{};a:=NewApp(f)
  if err:=a.Run([]string{"open",sel});err!=nil{t.Fatal(err)}
  last:=strings.Join(f.calls[len(f.calls)-1]," ")
  if !strings.Contains(last,"http://127.0.0.1:28080"){t.Fatalf("unexpected opener: %s",last)}
 }
}
func TestInvalidSelection(t *testing.T){
 for _,sel:=range []string{"0","7","missing"}{
  f:=&psRunner{};if err:=NewApp(f).Run([]string{"open",sel});err==nil{t.Fatalf("accepted %s",sel)}
 }
}
func TestDockerFailure(t *testing.T){
 if err:=NewApp(&psRunner{fail:true}).Run([]string{"ps"});err==nil{t.Fatal("expected failure")}
}
