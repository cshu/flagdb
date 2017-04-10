package main

import(
	"database/sql"
	"context"
//	"fmt"
	"log"
	"bytes"
	"encoding/binary"
	"net"
	"net/http"
	"io/ioutil"
	"os"
	_ "github.com/lib/pq"
//	"strings"
)


func main() {
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
		if r.Method=="GET"{
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.Write(htmlbytes)
			//?check err?
			return
		}
		//?only works for ipv4? remotestr:=r.RemoteAddr[:strings.IndexByte(r.RemoteAddr,':')]
		host,_,err:=net.SplitHostPort(r.RemoteAddr)
		if err!=nil{//?
			return
		}
		cliipbytes:=net.ParseIP(host)
		if cliipbytes==nil{//?
			return
		}
		if len(cliipbytes)==net.IPv4len{
			//undone
		}
		body,err:=ioutil.ReadAll(r.Body)
		if err!=nil{//?
			return
		}
		tr,err:=db.BeginTx(context.Background(),&sql.TxOptions{sql.LevelSerializable,false})
		if err!=nil{//?
			return
		}
		defer tr.Commit()

		rows,err:=tr.Query(`select i4,i6,d4,d6 from sp_urlrate where u=$1`,body[1:])
		if err!=nil{//?
			return
		}
		defer rows.Close()//?closing too late?
		var i4,i6,d4,d6 []byte
		rowsnext:=rows.Next()
		if rowsnext{
			rows.Scan(&i4,&i6,&d4,&d6)
		}
		switch body[0]{
		case 0:
			w.Header().Set("Content-Type", "application/octet-stream")
			ratebuf:=new(bytes.Buffer)
			binary.Write(ratebuf,binary.LittleEndian,int32(len(i4)/4+len(i6)/16-len(d4)/4-len(d6)/16))
			w.Write(ratebuf.Bytes())
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
							tr.Exec(`update sp_urlrate set d4=$1,i4=$2 where u=$3`,append(cod4[:len(cod4)-len(d4)],d4[net.IPv4len:]...),append(coi4,cliipbytes...),body[1:])//?check err?
							return
						}
						d4=d4[net.IPv4len:]
					}
					//optimize sort so each time you can do binary search?
					tr.Exec(`update sp_urlrate set i4=$1 where u=$2`,append(coi4,cliipbytes...),body[1:])//?check err?
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,i4)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						return
					}
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
							tr.Exec(`update sp_urlrate set d6=$1,i6=$2 where u=$3`,append(cod6[:len(cod6)-len(d6)],d6[net.IPv6len:]...),append(coi6,cliipbytes...),body[1:])//?check err?
							return
						}
						d6=d6[net.IPv6len:]
					}
					tr.Exec(`update sp_urlrate set i6=$1 where u=$2`,append(coi6,cliipbytes...),body[1:])//?check err?
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,i6)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						return
					}
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
							tr.Exec(`update sp_urlrate set i4=$1,d4=$2 where u=$3`,append(coi4[:len(coi4)-len(i4)],i4[net.IPv4len:]...),append(cod4,cliipbytes...),body[1:])//?check err?
							return
						}
						i4=i4[net.IPv4len:]
					}
					tr.Exec(`update sp_urlrate set d4=$1 where u=$2`,append(cod4,cliipbytes...),body[1:])//?check err?
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,d4)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						return
					}
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
							tr.Exec(`update sp_urlrate set i6=$1,d6=$2 where u=$3`,append(coi6[:len(coi6)-len(i6)],i6[net.IPv6len:]...),append(cod6,cliipbytes...),body[1:])//?check err?
							return
						}
						i6=i6[net.IPv6len:]
					}
					tr.Exec(`update sp_urlrate set d6=$1 where u=$2`,append(cod6,cliipbytes...),body[1:])//?check err?
					return
				}else{
					if _,err:=tr.Exec(`insert into sp_urlrate(u,d6)values($1,$2)`,body[1:],cliipbytes);err!=nil{
						return
					}
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
