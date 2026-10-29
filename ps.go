package main

import (
 "encoding/json"
 "errors"
 "fmt"
 "net"
 "net/url"
 "os"
 "runtime"
 "sort"
 "strconv"
 "strings"
 "text/tabwriter"
)

type dockerRow struct {
 ID string
 Names string
 Image string
 Command string
 Ports string
}
type listedContainer struct {
 dockerRow
 ImageID string
}
func (a *App) runningContainers() ([]listedContainer,error) {
 raw,err:=a.run.Run("docker","--context","colima","ps","--format","{{json .}}")
 if err!=nil{return nil,fmt.Errorf("docker ps: %w: %s",err,raw)}
 rows:=[]listedContainer{}
 for _,line:=range strings.Split(strings.TrimSpace(raw),"\n") {
  if strings.TrimSpace(line)==""{continue}
  var row dockerRow
  if err:=json.Unmarshal([]byte(line),&row);err!=nil{return nil,fmt.Errorf("decode docker ps output: %w",err)}
  if row.ID==""||row.Names==""{return nil,errors.New("incomplete docker ps response")}
  rows=append(rows,listedContainer{dockerRow:row})
 }
 sort.Slice(rows,func(i,j int)bool{return rows[i].Names<rows[j].Names})
 return rows,nil
}
func (a *App) ps(args []string) error {
 if len(args)!=0{return errors.New("usage: colima-cli ps")}
 rows,err:=a.runningContainers();if err!=nil{return err}
 w:=tabwriter.NewWriter(a.out,0,0,2,' ',0)
 fmt.Fprintln(w,"#NR\tNAME\tCONTAINERID\tURL/PORTS\tIMAGE ID\tCOMMAND")
 for i,row:=range rows {
  imageID:=row.Image
  inspected,e:=a.run.Run("docker","--context","colima","inspect","--format","{{.Image}}",row.ID)
  if e==nil&&strings.TrimSpace(inspected)!=""{imageID=strings.TrimSpace(inspected)}
  if strings.HasPrefix(imageID,"sha256:"){imageID=strings.TrimPrefix(imageID,"sha256:");if len(imageID)>12{imageID=imageID[:12]}}
  ports:=row.Ports
  if target,e:=containerURL(ports);e==nil {ports=target+" ("+ports+")"}
  fmt.Fprintf(w,"%d\t%s\t%s\t%s\t%s\t%s\n",i+1,row.Names,row.ID,ports,imageID,row.Command)
 }
 return w.Flush()
}
func containerURL(ports string)(string,error){
 for _,part:=range strings.Split(ports,","){
  part=strings.TrimSpace(part)
  if !strings.Contains(part,"->")||!strings.HasSuffix(part,"/tcp"){continue}
  hostPort:=strings.SplitN(part,"->",2)[0]
  host,port,e:=net.SplitHostPort(hostPort)
  if e!=nil {
   p:=strings.LastIndex(hostPort,":");if p<0{continue}
   host,port=hostPort[:p],hostPort[p+1:]
  }
  n,e:=strconv.Atoi(port);if e!=nil||n<1||n>65535{continue}
  host=strings.Trim(host,"[]")
  if host==""||host=="0.0.0.0"||host=="::"{host="localhost"}
  scheme:="http"
  if n==443||n==8443{scheme="https"}
  u:=url.URL{Scheme:scheme,Host:net.JoinHostPort(host,port)}
  return u.String(),nil
 }
 return "",errors.New("container has no published TCP port")
}
func (a *App) openContainer(args []string)error{
 if len(args)!=1{return errors.New("usage: colima-cli open <NR>|<NAME>")}
 rows,err:=a.runningContainers();if err!=nil{return err}
 var chosen *listedContainer
 if n,e:=strconv.Atoi(args[0]);e==nil {
  if n<1||n>len(rows){return fmt.Errorf("container number %d out of range",n)}
  chosen=&rows[n-1]
 } else {
  for i:=range rows {if rows[i].Names==args[0]{chosen=&rows[i];break}}
 }
 if chosen==nil{return fmt.Errorf("running container %q not found",args[0])}
 target,e:=containerURL(chosen.Ports);if e!=nil{return fmt.Errorf("%s: %w",chosen.Names,e)}
 opener:="open";if runtime.GOOS!="darwin"{opener="xdg-open"}
 if _,e:=os.Stat("/usr/bin/"+opener);e!=nil && runtime.GOOS=="darwin"{return fmt.Errorf("browser opener unavailable: %w",e)}
 result,e:=a.run.Run(opener,target)
 if e!=nil{return fmt.Errorf("open %s: %w: %s",target,e,result)}
 fmt.Fprintln(a.out,target)
 return nil
}
