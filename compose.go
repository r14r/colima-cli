package main

import (
 "errors"
 "flag"
 "fmt"
 "io"
 "os"
 "os/exec"
 "strings"
 "time"
)

type StreamRunner interface { Stream(io.Writer,io.Writer,string,...string) error }
func(systemRunner)Stream(out,err io.Writer,name string,args ...string)error{
 cmd:=exec.Command(name,args...)
 cmd.Stdout=out;cmd.Stderr=err;cmd.Stdin=os.Stdin
 return cmd.Run()
}
func(a *App)live(name string,args ...string)error{
 if s,ok:=a.run.(StreamRunner);ok{return s.Stream(a.out,a.err,name,args...)}
 return a.call(name,args...)
}
func(a *App)containers(args []string)error{
 fs:=flag.NewFlagSet("containers",flag.ContinueOnError);fs.SetOutput(a.err)
 running:=fs.Bool("running",false,"running containers only")
 stats:=fs.Bool("stats",false,"CPU/memory snapshot")
 if e:=fs.Parse(args);e!=nil{return e}
 if fs.NArg()>0{return errors.New("unexpected containers arguments")}
 if *stats{
  x:=[]string{"--context","colima","stats","--no-stream"}
  if !*running{x=append(x,"--all")}
  return a.call("docker",x...)
 }
 x:=[]string{"--context","colima","ps"}
 if !*running{x=append(x,"--all")}
 return a.call("docker",x...)
}
func(a *App)compose(args []string)error{
 if len(args)==0{return errors.New("compose requires subcommand")}
 cmd,rest:=args[0],args[1:]
 switch cmd{
 case "ls","ps","config":
  return a.call("docker",append([]string{"--context","colima","compose",cmd},rest...)...)
 case "logs","up","down":
  return a.live("docker",append([]string{"--context","colima","compose",cmd},rest...)...)
 case "stats":
  ids,e:=a.run.Run("docker","--context","colima","compose","ps","-q")
  if e!=nil{return fmt.Errorf("compose ps -q failed: %w: %s",e,ids)}
  containers:=strings.Fields(ids)
  if len(containers)==0{fmt.Fprintln(a.out,"No containers in current Compose project");return nil}
  return a.call("docker",append([]string{"--context","colima","stats","--no-stream"},containers...)...)
 default:return fmt.Errorf("unknown compose action %q",cmd)
 }
}
func(a *App)dashboard(args []string)error{
 fs:=flag.NewFlagSet("dashboard",flag.ContinueOnError);fs.SetOutput(a.err)
 interval:=fs.Int("interval",3,"refresh seconds")
 once:=fs.Bool("once",false,"one snapshot")
 if e:=fs.Parse(args);e!=nil{return e}
 if fs.NArg()>0||*interval<1||*interval>3600{return errors.New("invalid dashboard options: --interval must be 1..3600")}
 sections:=[]struct{title,name string;args []string}{
 {"VM status","colima",[]string{"status"}},
 {"VM CPU / memory","colima",[]string{"list"}},
 {"VM disk","colima",[]string{"ssh","--","df","-h","/"}},
 {"Containers","docker",[]string{"--context","colima","ps","--all"}},
 {"Resource usage","docker",[]string{"--context","colima","stats","--all","--no-stream"}},
 {"Compose projects","docker",[]string{"--context","colima","compose","ls","--all"}},
 }
 for{
  if !*once{fmt.Fprint(a.out,"\x1b[H\x1b[2J")}
  fmt.Fprintf(a.out,"COLIMA DASHBOARD %s (refresh %ds)\n",time.Now().Format("2006-01-02 15:04:05"),*interval)
  for _,s:=range sections{
   fmt.Fprintf(a.out,"\n== %s ==\n",s.title)
   out,e:=a.run.Run(s.name,s.args...)
   if e!=nil{fmt.Fprintf(a.out,"Unavailable: %v\n",e)}else{fmt.Fprintln(a.out,out)}
  }
  if *once{return nil}
  fmt.Fprintln(a.out,"\nCtrl+C to exit")
  time.Sleep(time.Duration(*interval)*time.Second)
 }
}
