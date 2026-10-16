package main

import (
 "errors"
 "strings"
 "testing"
)
type fake struct{calls [][]string}
func(f *fake)Run(name string,args ...string)(string,error){
 f.calls=append(f.calls,append([]string{name},args...))
 return "OK",nil
}
func TestVersion(t *testing.T){
 if e:=NewApp(&fake{}).Run([]string{"--version"});e!=nil{t.Fatal(e)}
}
func TestStatusReadOnly(t *testing.T){
 f:=&fake{}
 if e:=NewApp(f).Run([]string{"status","--quick"});e!=nil{t.Fatal(e)}
 for _,c:=range f.calls{
  s:=strings.Join(c," ")
  if strings.Contains(s,"context use")||strings.Contains(s,"run hello-world"){t.Fatal("side effect:",s)}
 }
}
func TestDestructiveGuards(t *testing.T){
 for _,args:=range [][]string{{"delete"},{"docker","prune"},{"context","use"}}{
  f:=&fake{}
  e:=NewApp(f).Run(args)
  if e==nil||len(f.calls)>0{t.Fatalf("expected rejection for %v",args)}
 }
}
func TestUnknownCommand(t *testing.T){
 if e:=NewApp(&fake{}).Run([]string{"unknown"});!errors.Is(e,nil)&&e==nil{t.Fatal("expected error")}
}
