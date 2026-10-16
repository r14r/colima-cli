package main
import("strings";"testing")
func TestContainers(t *testing.T){
 for _,tt:=range []struct{args []string;want string}{
  {[]string{"containers"},"docker --context colima ps --all"},
  {[]string{"containers","--running"},"docker --context colima ps"},
  {[]string{"containers","--stats"},"docker --context colima stats --no-stream --all"},
 }{
  f:=&fake{};a:=app(f);defer a.out.Close()
  if e:=a.Run(tt.args);e!=nil{t.Fatal(e)}
  if got:=strings.Join(f.calls[0]," ");got!=tt.want{t.Fatalf("%q != %q",got,tt.want)}
 }
}
func TestCompose(t *testing.T){
 for _,tt:=range []struct{args []string;want string}{
  {[]string{"compose","ls"},"docker --context colima compose ls"},
  {[]string{"compose","ps","-a"},"docker --context colima compose ps -a"},
  {[]string{"compose","logs","-f"},"docker --context colima compose logs -f"},
  {[]string{"compose","up","-d","--build"},"docker --context colima compose up -d --build"},
  {[]string{"compose","down"},"docker --context colima compose down"},
  {[]string{"compose","config"},"docker --context colima compose config"},
 }{
  f:=&fake{};a:=app(f);defer a.out.Close()
  if e:=a.Run(tt.args);e!=nil{t.Fatal(e)}
  if got:=strings.Join(f.calls[0]," ");got!=tt.want{t.Fatalf("%q != %q",got,tt.want)}
 }
}
func TestComposeStats(t *testing.T){
 f:=&fake{};a:=app(f);defer a.out.Close()
 if e:=a.Run([]string{"compose","stats"});e!=nil{t.Fatal(e)}
 if len(f.calls)!=2{t.Fatalf("expected 2 calls got %d",len(f.calls))}
}
func TestDashboardOnce(t *testing.T){
 f:=&fake{};a:=app(f);defer a.out.Close()
 if e:=a.Run([]string{"dashboard","--once"});e!=nil{t.Fatal(e)}
 if len(f.calls)!=6{t.Fatalf("expected six probes got %d",len(f.calls))}
}
