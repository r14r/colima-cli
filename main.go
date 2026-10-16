package main

import (
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "os"
 "os/exec"
 "strconv"
 "strings"
 "time"
)

const version = "0.0.1"

type Runner interface { Run(string,...string)(string,error) }
type systemRunner struct{}
func (systemRunner) Run(name string,args ...string)(string,error) {
 b,e:=exec.Command(name,args...).CombinedOutput()
 return strings.TrimSpace(string(b)),e
}
type App struct {run Runner}
func NewApp(r Runner)*App{return &App{r}}
func (a *App) call(name string,args ...string) error {
 out,e:=a.run.Run(name,args...)
 if out!="" {fmt.Println(out)}
 if e!=nil {fmt.Fprintf(os.Stderr,"%s %s: %v\n",name,strings.Join(args," "),e)}
 return e
}
func (a *App) resources()error {
 for _,s:=range [][]string{{"df","-h","/"},{"free","-h"},{"nproc"}} {
  if e:=a.call("colima",append([]string{"ssh","--"},s...)...);e!=nil{return e}
 }
 return nil
}
func (a *App) docker()error {
 if e:=a.call("docker","context","ls");e!=nil{return e}
 return a.call("docker","--context","colima","info")
}
func (a *App) status(args []string)error {
 fs:=flag.NewFlagSet("status",flag.ContinueOnError)
 quick:=fs.Bool("quick",false,"fast check")
 jsonOut:=fs.Bool("json",false,"JSON status")
 resources:=fs.Bool("resources",false,"resource usage")
 docker:=fs.Bool("docker",false,"Docker details")
 if e:=fs.Parse(args);e!=nil{return e}
 c,e:=a.run.Run("colima","status")
 running:=e==nil
 d,de:=a.run.Run("docker","--context","colima","info","--format","{{.ServerVersion}}")
 healthy:=de==nil
 if *jsonOut {
  v:=map[string]interface{}{"version":version,"colima_running":running,"docker_colima_available":healthy,"colima_status":c,"docker_server_version":d}
  b,_:=json.MarshalIndent(v,"","  ")
  fmt.Println(string(b));return nil
 }
 fmt.Println("COLIMA STATUS")
 if c!=""{fmt.Println(c)}
 if !running{return errors.New("Colima is unavailable")}
 if *resources{return a.resources()}
 if *docker{return a.docker()}
 fmt.Printf("Docker daemon: %v (%s)\n",healthy,d)
 if *quick{return nil}
 fmt.Println("\nInstances:");a.call("colima","list")
 fmt.Println("\nDisk:");a.call("colima","ssh","--","df","-h","/")
 fmt.Println("\nMemory:");a.call("colima","ssh","--","free","-h")
 fmt.Println("\nCPUs:");a.call("colima","ssh","--","nproc")
 fmt.Println("\nContexts:");a.call("docker","context","ls")
 return nil
}
func (a *App) doctor()error {
 fail:=false
 for _,n:=range []string{"colima","docker","limactl"} {
  _,e:=exec.LookPath(n);fmt.Printf("%-10s %v\n",n,e==nil)
  if e!=nil{fail=true}
 }
 if _,e:=a.run.Run("colima","status");e!=nil{fmt.Println("Colima VM: unavailable");fail=true}
 if _,e:=a.run.Run("docker","--context","colima","info");e!=nil{fmt.Println("Colima Docker daemon: unavailable");fail=true}
 if fail{return errors.New("doctor checks failed")}
 fmt.Println("All checks passed");return nil
}
func has(args []string,value string)bool{for _,s:=range args{if s==value{return true}};return false}
func without(args []string,value string)[]string{r:=[]string{};for _,s:=range args{if s!=value{r=append(r,s)}};return r}
func usage(){fmt.Print(`colima-cli 0.0.1 — Colima and Docker management

USAGE
  colima-cli <command> [options]

DIAGNOSTICS
  status [--quick|--resources|--docker|--json]
  doctor
  test
  resources
  list
  logs [args...]
  version
  whatsnew

VM MANAGEMENT
  start [colima flags]
  stop [colima flags]
  restart [colima flags]
  delete --yes [flags]
  ssh [args...]
  exec -- <command>
  template [args...]

DOCKER
  docker info
  docker ps [args...]
  docker images
  docker df
  docker prune --yes
  context list
  context current
  context use --yes
  wait [seconds]

Mutating context changes and destructive operations require --yes.
`)}
func (a *App) Run(args []string)error{
 if len(args)==0{usage();return nil}
 cmd,rest:=args[0],args[1:]
 switch cmd {
 case "help","--help","-h":usage();return nil
 case "version","--version","-V":fmt.Println("colima-cli",version);return nil
 case "whatsnew":fmt.Println("0.0.1: Initial release — status, diagnostics, VM lifecycle and Docker helpers.");return nil
 case "status":return a.status(rest)
 case "doctor":return a.doctor()
 case "test","selftest":return a.call("docker","--context","colima","run","--rm","hello-world")
 case "resources":return a.resources()
 case "list","logs","start","stop","restart","template","ssh":return a.call("colima",append([]string{cmd},rest...)...)
 case "exec":
  if len(rest)>0&&rest[0]=="--"{rest=rest[1:]}
  if len(rest)==0{return errors.New("exec expects a command")}
  return a.call("colima",append([]string{"ssh","--"},rest...)...)
 case "delete":
  if !has(rest,"--yes"){return errors.New("delete requires --yes")}
  return a.call("colima",append([]string{"delete"},without(rest,"--yes")...)...)
 case "context":
  if len(rest)==0{return errors.New("context expects list/current/use")}
  switch rest[0]{
  case "list":return a.call("docker","context","ls")
  case "current":return a.call("docker","context","show")
  case "use":
   if !has(rest[1:],"--yes"){return errors.New("context use requires --yes")}
   return a.call("docker","context","use","colima")
  }
  return errors.New("unknown context action")
 case "docker":
  if len(rest)==0{return a.docker()}
  switch rest[0] {
  case "info":return a.call("docker","--context","colima","info")
  case "ps":return a.call("docker",append([]string{"--context","colima","ps"},rest[1:]...)...)
  case "images":return a.call("docker","--context","colima","images")
  case "df":return a.call("docker","--context","colima","system","df")
  case "prune":
   if !has(rest[1:],"--yes"){return errors.New("docker prune requires --yes")}
   return a.call("docker","--context","colima","system","prune","--force")
  }
  return errors.New("unknown Docker action")
 case "wait":
  timeout:=30*time.Second
  if len(rest)>0{n,e:=strconv.Atoi(rest[0]);if e!=nil||n<1{return errors.New("wait expects positive seconds")};timeout=time.Duration(n)*time.Second}
  deadline:=time.Now().Add(timeout)
  for time.Now().Before(deadline){
   if _,e:=a.run.Run("docker","--context","colima","info");e==nil{fmt.Println("Colima Docker ready");return nil}
   time.Sleep(time.Second)
  }
  return errors.New("timed out waiting for Colima")
 }
 return fmt.Errorf("unknown command %q (try --help)",cmd)
}
func main(){if e:=NewApp(systemRunner{}).Run(os.Args[1:]);e!=nil{fmt.Fprintln(os.Stderr,"Error:",e);os.Exit(1)}}
