package main

import(
	"database/sql"
//	"context"
//	"fmt"
	"log"
	"bytes"
	"encoding/binary"
	"net"
	"net/http"
	"io/ioutil"
	"os"
	"strings"
	_ "github.com/lib/pq"
)


func main() {
	log.SetFlags(log.LstdFlags|log.Lshortfile)
	htmlbytes,err:=ioutil.ReadFile("index.htm")
	if err!=nil{
		log.Fatalln(err)
	}
	dburl:=os.Getenv("DATABASE_URL")
	//remove?
	//connection, _ := pq.ParseURL(url)
	//connection += " sslmode=require"
	db,err:=sql.Open("postgres",dburl)
	if err!=nil{
		log.Fatalln(err)
	}
	if _,err:=db.Exec("CREATE TABLE IF NOT EXISTS sp_urlrate(u bytea,i4 bytea,i6 bytea,d4 bytea,d6 bytea)"); err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		//log.Println("RemoteAddr: "+r.RemoteAddr)//remove
		//log.Println(r.Header.Get("X-Forwarded-For"))//remove
		if r.Method=="GET"{
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.Write(htmlbytes)
			//?check err?
			return
		}
		var cliipbytes net.IP
		var err error
		if host:=r.Header.Get("X-Forwarded-For");host==""{
			//remotestr:=r.RemoteAddr[:strings.IndexByte(r.RemoteAddr,':')]//?only works for ipv4?
			host,_,err=net.SplitHostPort(r.RemoteAddr)
			if err!=nil{//?
				log.Println(err)
				return
			}
			cliipbytes=net.ParseIP(host)
			if cliipbytes==nil{//?
				log.Println("ParseIP returns nil")
				return
			}
		}else{
			for _,hoste:=range strings.Split(host,","){//you should only need the left-most ip in X-Forwarded-For, but heroku doesn't conform? (quirk?)
				if successhost:=net.ParseIP(strings.TrimSpace(hoste));successhost!=nil{
					cliipbytes=successhost
					if cliipbytes.IsGlobalUnicast(){
						goto startreading
					}
				}
			}
			if cliipbytes==nil{return}
		}
		startreading:
		log.Println("Remote IP: "+cliipbytes.String())//remove
		body,err:=ioutil.ReadAll(r.Body)
		if err!=nil{//?
			log.Println(err)
			return
		}
		if len(body)<2{return}
		for _,bodybyte := range body[1:]{
			if bodybyte==0x2f || (bodybyte>0x40 && bodybyte<0x5b){//upper case not allowed, checked as UTF-8
				return
			}
		}
		//tr,err:=db.BeginTx(context.Background(),&sql.TxOptions{sql.LevelSerializable,false})//? got error! not supported?
		tr,err:=db.Begin()
		if err!=nil{//?
			log.Println(err)
			return
		}
		defer tr.Commit()
		_,err=tr.Exec(`LOCK TABLE sp_urlrate IN ACCESS EXCLUSIVE MODE`)
		if err!=nil{//?
			log.Println(err)
			return
		}

		rows,err:=tr.Query(`select i4,i6,d4,d6 from sp_urlrate where u=$1`,body[1:])
		if err!=nil{//?
			log.Println(err)
			return
		}
		//defer rows.Close()//note must call rows.Close before next query otherwise there will be error?
		var rofhostnm int
		var i4,i6,d4,d6 []byte
		rowsnext:=rows.Next()
		if rowsnext{
			rows.Scan(&i4,&i6,&d4,&d6)
			rofhostnm=len(i4)/4+len(i6)/16-len(d4)/4-len(d6)/16
		}
		rows.Close()//? no panic allowed before rows.Close() since you are not using defer?
		w.Header().Set("Content-Type", "application/octet-stream")
		ratebuf:=new(bytes.Buffer)
		switch body[0]{
		case 0:
			binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
			w.Write(ratebuf.Bytes())
			return
		case 1:
			if len(cliipbytes)==net.IPv4len{
				if rowsnext{
					coi4:=i4
					for len(i4)!=0 {
						if bytes.HasPrefix(i4,cliipbytes){
							return
						}
						i4=i4[net.IPv4len:]
					}
					cod4:=d4
					for len(d4)!=0 {
						if bytes.HasPrefix(d4,cliipbytes){
							if _,err:=tr.Exec(`update sp_urlrate set d4=$1,i4=$2 where u=$3`,append(cod4[:len(cod4)-len(d4)],d4[net.IPv4len:]...),append(coi4,cliipbytes...),body[1:]);err!=nil{
								log.Println(err)
								return
							}
							rofhostnm+=2
							binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
							w.Write(ratebuf.Bytes())
							return
						}
						d4=d4[net.IPv4len:]
					}
					//optimize sort so each time you can do binary search?
					if _,err:=tr.Exec(`update sp_urlrate set i4=$1 where u=$2`,append(coi4,cliipbytes...),body[1:]);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm++
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,i4)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm++
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}
			}else if len(cliipbytes)==net.IPv6len{//?check cuz you are paranoid
				if rowsnext{
					coi6:=i6
					for len(i6)!=0 {
						if bytes.HasPrefix(i6,cliipbytes){
							return
						}
						i6=i6[net.IPv6len:]
					}
					cod6:=d6
					for len(d6)!=0 {
						if bytes.HasPrefix(d6,cliipbytes){
							if _,err:=tr.Exec(`update sp_urlrate set d6=$1,i6=$2 where u=$3`,append(cod6[:len(cod6)-len(d6)],d6[net.IPv6len:]...),append(coi6,cliipbytes...),body[1:]);err!=nil{
								log.Println(err)
								return
							}
							rofhostnm+=2
							binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
							w.Write(ratebuf.Bytes())
							return
						}
						d6=d6[net.IPv6len:]
					}
					if _,err:=tr.Exec(`update sp_urlrate set i6=$1 where u=$2`,append(coi6,cliipbytes...),body[1:]);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm++
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,i6)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm++
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}
			}else{return}
		case 2:
			if len(cliipbytes)==net.IPv4len{
				if rowsnext{
					cod4:=d4
					for len(d4)!=0 {
						if bytes.HasPrefix(d4,cliipbytes){
							return
						}
						d4=d4[net.IPv4len:]
					}
					coi4:=i4
					for len(i4)!=0 {
						if bytes.HasPrefix(i4,cliipbytes){
							if _,err:=tr.Exec(`update sp_urlrate set i4=$1,d4=$2 where u=$3`,append(coi4[:len(coi4)-len(i4)],i4[net.IPv4len:]...),append(cod4,cliipbytes...),body[1:]);err!=nil{
								log.Println(err)
								return
							}
							rofhostnm-=2
							binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
							w.Write(ratebuf.Bytes())
							return
						}
						i4=i4[net.IPv4len:]
					}
					if _,err:=tr.Exec(`update sp_urlrate set d4=$1 where u=$2`,append(cod4,cliipbytes...),body[1:]);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm--
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,d4)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm--
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}
			}else if len(cliipbytes)==net.IPv6len{//?check cuz you are paranoid
				if rowsnext{
					cod6:=d6
					for len(d6)!=0 {
						if bytes.HasPrefix(d6,cliipbytes){
							return
						}
						d6=d6[net.IPv6len:]
					}
					coi6:=i6
					for len(i6)!=0 {
						if bytes.HasPrefix(i6,cliipbytes){
							if _,err:=tr.Exec(`update sp_urlrate set i6=$1,d6=$2 where u=$3`,append(coi6[:len(coi6)-len(i6)],i6[net.IPv6len:]...),append(cod6,cliipbytes...),body[1:]);err!=nil{
								log.Println(err)
								return
							}
							rofhostnm-=2
							binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
							w.Write(ratebuf.Bytes())
							return
						}
						i6=i6[net.IPv6len:]
					}
					if _,err:=tr.Exec(`update sp_urlrate set d6=$1 where u=$2`,append(cod6,cliipbytes...),body[1:]);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm--
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,d6)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						log.Println(err)
						return
					}
					rofhostnm--
					binary.Write(ratebuf,binary.LittleEndian,int32(rofhostnm))
					w.Write(ratebuf.Bytes())
					return
				}
			}else{return}
		}
		//_,err:=
		//w.Write([]byte(host))
		//if err!=nil{
		//	log.Println(err)
		//	return
		//}
	})
	log.Println(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
