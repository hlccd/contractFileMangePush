package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/gomail-2"
	Database "hlccd"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func toString(s string) string {
	return "\"" + s + "\""
}

type AccessToken struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}
type Err struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}
type remind struct {
	ContractName string `json:"contractName"`
	Name         string `json:"name"`
	Time         int64  `json:"time"`
	Customer     string `json:"customer"`
	Uid          string `json:"uid"`
	Contract     string `json:"contract"`
	Touser       string `json:"touser"`
	Email        string `json:"email"`
}

func CreateList() (Database.DB, []Database.Table, *sql.DB) {
	//数据库登陆结构信息
	D := Database.DB{
		DriverName: "mysql",
		User:       "root",
		Password:   "2975hLcCd",
		Tcp:        "localhost:3306",
		Name:       "hlccd",
	}
	Tab := make([]Database.Table, 0)
	//0号数据表：用户总列表
	Tab = append(Tab,
		Database.Table{
			Name: "account_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"openid char(30)",
				"uid bigint",
				"name varchar(30)",
				"power int",
				"department int",
			},
			Annotation: "auto_increment=99999",
		})
	//1号数据表：部门总列表
	Tab = append(Tab,
		Database.Table{
			Name: "department_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"name varchar(100)",
			},
			Annotation: "auto_increment=1",
		})
	//2号数据表：合同总列表
	Tab = append(Tab,
		Database.Table{
			Name: "contract_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"name varchar(100)",
				"notes varchar(100)",
				"uploader bigint",
				"department int",
				"pdfType varchar(10)",
			},
			Annotation: "auto_increment=1",
		})
	//3号数据表：图片总列表
	Tab = append(Tab,
		Database.Table{
			Name: "img_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"contract bigint",
				"imgType varchar(10)",
			},
			Annotation: "auto_increment=1",
		})
	//4号数据表：头像总列表
	Tab = append(Tab,
		Database.Table{
			Name: "head_list",
			Value: []string{
				"account bigint",
				"imgType varchar(10)",
			},
			Annotation: "",
		})
	//5号数据表：时间提醒总列表
	Tab = append(Tab,
		Database.Table{
			Name: "time_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"name varchar(100)",
				"contract bigint",
				"timestamp bigint",
			},
			Annotation: "auto_increment=1",
		})
	//5号数据表：订阅总表
	Tab = append(Tab,
		Database.Table{
			Name: "subscribe_list",
			Value: []string{
				"uid bigint",
				"subscribe int",
			},
			Annotation: "",
		})
	//启动数据库
	db, b, _ := Database.OpenDatabase(D)
	if b {
		//启动成功后创新所有Tab中的数据表
		for x := range Tab {
			Database.CreateTable(db, Tab[x])
		}
	}
	return D, Tab, db
}

func toTimestamp0(s string) int64 {
	timeLayout := "2006-01-02 15:04:05"
	times, _ := time.Parse(timeLayout, s[:10]+" 00:00:00")
	return times.Unix()
}
func toTimestamp8(s string) int64 {
	timeLayout := "2006-01-02 15:04:05"
	times, _ := time.Parse(timeLayout, s[:10]+" 10:00:00")
	return times.Unix()
}
func toTimeS(i int64) string {
	timeLayout := "2006-01-02 15:04:05"
	datetime := time.Unix(i, 0).Format(timeLayout)
	return datetime[:10]
}

func getAccessToken() string {
	url := "https://api.weixin.qq.com/cgi-bin/token" +
		"?grant_type=client_credential" +
		"&appid=wx08c2e30623d360e9" +
		"&secret=c9875679121476cdbf99f09c944b5122"
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	var token AccessToken
	json.Unmarshal([]byte(string(body)), &token)
	return token.Access_token
}

type Contract struct {
	customer     string
	contractName string
	beginTime    string
	productName  string
	endTime      string
}

func subscribe(c Contract, access_token string, touser string, miniprogram_state string) (E Err) {
	template_id := "L5mwCrjKD3yKItp3S7ASDb_ErLUc2NqHoRj5kMwd8iA"
	s := "{" + toString("touser") + ":" + toString(touser) + "," +
		toString("template_id") + ":" + toString(template_id) + "," +
		toString("miniprogram_state") + ":" + toString(miniprogram_state) + "," +
		toString("data") + ":{" +
		toString("thing1") + ":{" + toString("value") + ":" + toString(c.customer) + "}," +
		toString("thing2") + ":{" + toString("value") + ":" + toString(c.contractName) + "}," +
		toString("time4") + ":{" + toString("value") + ":" + toString(c.beginTime) + "}," +
		toString("thing3") + ":{" + toString("value") + ":" + toString(c.productName) + "}," +
		toString("time5") + ":{" + toString("value") + ":" + toString(c.endTime) + "}" +
		"}}"
	var jsonStr = []byte(s)

	url := "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?" +
		"access_token=" + access_token
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	json.Unmarshal([]byte(string(body)), &E)
	return E
}
func PostSubscribe(touser string, c Contract) {
	access_token := getAccessToken()
	miniprogram_state := "trial"
	err := subscribe(c, access_token, touser, miniprogram_state)
	if err.Errcode != 0 {
		fmt.Println(err.Errmsg)
	}
}

func SendMail(mailTo []string, subject string, body string) error {
	mailConn := map[string]string{
		"user": "hlccd2975@163.com",
		"pass": "****",
		"host": "smtp.163.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(mailConn["user"], "合同提醒助手")) //这种方式可以添加别名，即“XX官方”
	//说明：如果是用网易邮箱账号发送，以下方法别名可以是中文，如果是qq企业邮箱，以下方法用中文别名，会报错，需要用上面此方法转码
	//m.SetHeader("From", "FB Sample"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“FB Sample”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	//m.SetHeader("From", mailConn["user"])
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	m.SetBody("text/html", body)    //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err

}
func email(user string, title string, body string) {
	//定义收件人
	mailTo := []string{
		user,
	}
	err := SendMail(mailTo, title, body)
	if err != nil {
		log.Println(err)
		fmt.Println("send fail")
		return
	}

	fmt.Println("send successfully")
}
func getList(db *sql.DB, time1 int64, time2 int64) (list []remind) {
	rows, _, _ := Database.SelectAllData(db, "time_list", "name,timestamp,contract", "timestamp>="+strconv.FormatInt(time1, 10)+" and timestamp<="+strconv.FormatInt(time2, 10))
	defer rows.Close()
	for rows.Next() {
		var name, contract string
		var timestamp int64
		err := rows.Scan(&name, &timestamp, &contract)
		contractName := Database.SelectKeyGetFieldS(db, "contract_list", "name", "id="+contract)
		customerID := Database.SelectKeyGetFieldS(db, "contract_list", "uploader", "id="+contract)
		customer := Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+customerID)
		uid := Database.SelectKeyGetFieldS(db, "account_list", "uid", "id="+customerID)
		touser := Database.SelectKeyGetFieldS(db, "account_list", "openid", "id="+customerID)
		email := Database.SelectKeyGetFieldS(db, "account_list", "ding", "id="+customerID)
		if err != nil {
			log.Fatal(err)
		}
		if name == "begin" {
			name = "合同起效"
		}
		if name == "deadline" {
			name = "合同过期"
		}
		list = append(list, remind{
			ContractName: contractName,
			Name:         name,
			Time:         timestamp,
			Customer:     customer,
			Uid:          uid,
			Contract:     contract,
			Touser:       touser,
			Email:        email,
		})
	}
	return list
}
func getList2(db *sql.DB, now int64) (list []remind) {
	rows, _, _ := Database.SelectAllData(db, "time_list", "timestamp,contract", "name='deadline' and timestamp>="+strconv.FormatInt(now, 10))
	defer rows.Close()
	for rows.Next() {
		var contract string
		var timestamp int64
		err := rows.Scan(&timestamp, &contract)
		if err != nil {
			log.Fatal(err)
		}
		if now>=Database.SelectKeyGetFieldI(db,"time_list","timestamp","name='begin' and id="+contract) {
			contractName := Database.SelectKeyGetFieldS(db, "contract_list", "name", "id="+contract)
			customerID := Database.SelectKeyGetFieldS(db, "contract_list", "uploader", "id="+contract)
			customer := Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+customerID)
			uid := Database.SelectKeyGetFieldS(db, "account_list", "uid", "id="+customerID)
			touser := Database.SelectKeyGetFieldS(db, "account_list", "openid", "id="+customerID)
			email := Database.SelectKeyGetFieldS(db, "account_list", "ding", "id="+customerID)
			list = append(list, remind{
				ContractName: contractName,
				Time:         timestamp,
				Customer:     customer,
				Uid:          uid,
				Contract:     contract,
				Touser:       touser,
				Email:        email,
			})
		}
	}
	return list
}
func push(db *sql.DB,list []remind,s string){
	for x := range list {
		beginI := Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+list[x].Contract+" and name='begin'")
		begin := toTimeS(beginI)
		deadlineI := Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+list[x].Contract+" and name='deadline'")
		deadline := toTimeS(deadlineI)
		c := Contract{
			customer:     list[x].Customer,
			contractName: list[x].ContractName,
			beginTime:    begin,
			productName:  list[x].Name,
			endTime:      deadline,
		}
		email(list[x].Email,list[x].ContractName+"合同将于"+s+"到达"+list[x].Name+"时间点",list[x].ContractName+"合同将于明日到期" +
			"\n合同签订人:"+list[x].Customer+
			"\n合同名称:"+list[x].ContractName+
			"\n合同开始时间:"+begin+
			"\n到期类型:"+list[x].Name+
			"\n合同截至时间"+deadline+
			"请注意合同时间以避免耽误相关事宜")
		PostSubscribe(list[x].Touser, c)
		Database.UpdateData(db, "subscribe_list", "subscribe=0", "uid='"+list[x].Uid+"'")
	}
}
func push2(db *sql.DB,list []remind){
	for x := range list {
		beginI := Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+list[x].Contract+" and name='begin'")
		begin := toTimeS(beginI)
		deadlineI := Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+list[x].Contract+" and name='deadline'")
		deadline := toTimeS(deadlineI)
		c := Contract{
			customer:     list[x].Customer,
			contractName: list[x].ContractName,
			beginTime:    begin,
			productName:  list[x].Name,
			endTime:      deadline,
		}
		if len(list[x].Email)>7 {
			email(list[x].Email,list[x].ContractName+"合同仍在有效期内",list[x].ContractName+"合同仍在有效期内" +
				"\n合同签订人:"+list[x].Customer+
				"\n合同名称:"+list[x].ContractName+
				"\n合同开始时间:"+begin+
				"\n到期类型:"+list[x].Name+
				"\n合同截至时间"+deadline+
				"请注意合同时间以避免耽误相关事宜")
			PostSubscribe(list[x].Touser, c)
			Database.UpdateData(db, "subscribe_list", "subscribe=0", "uid='"+list[x].Uid+"'")
		}
	}
}
func main() {
	_, _, db := CreateList()
	for {
		gap := int64(86400)
		nowTime0 := time.Now().UTC().Unix()
		nextTime1 := nowTime0 + gap
		nextTime2 := nextTime1 + gap
		nextTime3 := nextTime2 + gap
		nextTime4 := nextTime3 + gap
		list3 := getList(db, nextTime3, nextTime4)
		push(db,list3,"三日后")
		list1 := getList(db, nextTime1, nextTime2)
		push(db,list1,"明日")
		list0 := getList(db, nowTime0, nextTime1)
		push(db,list0,"今日")
		if time.Now().Day() == 24 {
			list:=getList2(db,nowTime0)
			push2(db,list)
		}
		nowT := time.Now().UTC().Unix()
		t := toTimestamp8(toTimeS(time.Now().UTC().Unix())) - nowT - gap/3
		if t < 0 {
			t += gap
		}
		fmt.Println(t)
		time.Sleep(time.Duration(t) * time.Second)
	}
}