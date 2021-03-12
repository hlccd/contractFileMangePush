package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	Database "hlccd"
	"hlccd/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type wx struct {
	Session_key string `json:"session_key"`
	Openid      string `json:"openid"`
}
type remind struct {
	Name string `json:"name"`
	Time string `json:"time"`
}
type department struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}
type contract struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Notes    string   `json:"notes"`
	Type     string   `json:"type"`
	Uploader string   `json:"uploader"`
	Begin    string   `json:"begin"`
	Deadline string   `json:"deadline"`
	Remind   []remind `json:"remind"`
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
				"ding char(50)",
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
				"type varchar(100)",
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
func ChangeToChinese(s string) string {
	if strings.Contains(s, "\\u") {
		sUnicodev := strings.Split(s, "\\u")
		var context string
		for _, v := range sUnicodev {
			if len(v) < 1 {
				continue
			}
			temp, err := strconv.ParseInt(v, 16, 32)
			if err != nil {
				panic(err)
			}
			context += fmt.Sprintf("%c", temp)
		}
		s = context
	}
	return s
}
func toTimestamp(s string) int64 {
	timeLayout := "2006-01-02 15:04:05"
	times, _ := time.Parse(timeLayout, s[:10]+" 12:00:00")
	return times.Unix()
}
func toTimeS(i int64) string {
	timeLayout := "2006-01-02 15:04:05"
	datetime := time.Unix(i, 0).Format(timeLayout)
	return datetime[:10]
}

//后台系统
//后台添加账号
func BackstageAdd(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/backstage/add", func(c *gin.Context) {
		id := int64(0)
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		openid := ""
		ding := ""
		nameS := ""
		power := int64(-1)
		department := 0
		tokenS := ""

		if uid == Database.SelectKeyGetFieldI(db, TabName, "uid", "uid="+uidS) {
			c.JSON(203, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       203,
				"message":    "账号已存在",
			})
		} else {
			Database.InsertData(db, TabName, "0,'',"+uidS+",'','',-1,0")
			Database.InsertData(db, "subscribe_list", uidS+",0")
			stmt, _, _ := Database.SelectLastData(db, "*", TabName)
			defer stmt.Close()
			if stmt.Next() {
				err := stmt.Scan(&id, &openid, &uid, &ding, &nameS, &power, &department)
				if err != nil {
					log.Fatal(err)
				}
				tokenS = token.CreateToken(strconv.FormatInt(id, 10))
				c.JSON(200, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": department,
					"token":      tokenS,
					"code":       200,
					"message":    "success",
				})
			}
		}
	})
}

//后台删除账号
func BackstageDelete(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/backstage/delete", func(c *gin.Context) {
		id := int64(0)
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		ding := ""
		openid := ""
		nameS := ""
		power := int64(-1)
		department := 0
		tokenS := ""
		if uid == Database.SelectKeyGetFieldI(db, TabName, "uid", "uid="+uidS) {
			Database.DeleteData(db, TabName, "uid="+uidS)
			Database.DeleteData(db, "subscribe_list", "uid="+uidS)
			c.JSON(200, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       200,
				"message":    "success",
			})
		} else {
			c.JSON(204, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       204,
				"message":    "删除失败,账号不存在",
			})
		}
	})
}

//后台设置权限
func BackstageSet(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/backstage/set", func(c *gin.Context) {
		id := int64(0)
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		ding := ""
		openid := ""
		nameS := ""
		powerS := c.Query("power")
		power, _ := strconv.ParseInt(powerS, 10, 64)
		department := 0
		tokenS := ""

		if uid != Database.SelectKeyGetFieldI(db, TabName, "uid", "uid="+uidS) {
			c.JSON(204, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       204,
				"message":    "查无此人",
			})
		} else {
			_, _ = Database.UpdateData(db, TabName, "power="+powerS, "uid="+uidS)
			stmt, _, _ := Database.SelectAllData(db, TabName, "*", "uid="+uidS)
			defer stmt.Close()
			if stmt.Next() {
				err := stmt.Scan(&id, &openid, &uid, &ding, &nameS, &power, &department)
				if err != nil {
					log.Fatal(err)
				}
				tokenS = token.CreateToken(strconv.FormatInt(id, 10))
				c.JSON(200, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": department,
					"token":      tokenS,
					"code":       200,
					"message":    "success",
				})
			}
		}
	})
}

//后台添加部门
func BackstageAddDepartment(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/backstage/department/add", func(c *gin.Context) {
		id := int64(0)
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)
		Database.InsertData(db, "department_list", "0,'"+nameS+"'")
		stmt, _, _ := Database.SelectLastData(db, "*", "department_list")
		defer stmt.Close()
		if stmt.Next() {
			err := stmt.Scan(&id, &nameS)
			if err != nil {
				log.Fatal(err)
			}
			c.JSON(200, gin.H{
				"id":      id,
				"name":    nameS,
				"code":    200,
				"message": "success",
			})
		}

	})
}

//后台权限系统
func BackstageSystem(r *gin.Engine, db *sql.DB, TabName string) {
	BackstageAdd(r, TabName, db)
	BackstageDelete(r, TabName, db)
	BackstageSet(r, TabName, db)
	BackstageAddDepartment(r, TabName, db)
}

//账号系统
//超管添加账号
func AccountAdd(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/account/add", func(c *gin.Context) {
		id := int64(0)
		idS := "0"
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		openid := ""
		ding := ""
		nameS := ""
		power := int64(-1)
		department := 0
		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      "",
				"code":       201,
				"message":    "token解析失败",
			})
		} else {
			idS = p.UserID
			id, _ = strconv.ParseInt(idS, 10, 64)
			if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+idS) != 0 {
				c.JSON(202, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": department,
					"token":      "",
					"code":       202,
					"message":    "非超管无权限",
				})
			} else {
				if uid == Database.SelectKeyGetFieldI(db, TabName, "uid", "uid="+uidS) {
					c.JSON(203, gin.H{
						"id":         id,
						"uid":        uid,
						"ding":       ding,
						"name":       nameS,
						"power":      power,
						"department": department,
						"token":      "",
						"code":       203,
						"message":    "账号已存在",
					})
				} else {
					Database.InsertData(db, TabName, "0,'',"+uidS+",'','',-1,0")
					Database.InsertData(db, "subscribe_list", uidS+",0")
					stmt, _, _ := Database.SelectAllData(db, TabName, "*", "uid="+uidS)
					defer stmt.Close()
					if stmt.Next() {
						err := stmt.Scan(&id, &openid, &uid, &ding, &nameS, &power, &department)
						if err != nil {
							log.Fatal(err)
						}
						tokenS = token.CreateToken(strconv.FormatInt(id, 10))
						c.JSON(200, gin.H{
							"id":         id,
							"openid":     openid,
							"uid":        uid,
							"ding":       ding,
							"name":       nameS,
							"power":      power,
							"department": department,
							"token":      tokenS,
							"code":       200,
							"message":    "success",
						})
					}
				}
			}
		}
	})
}

//code登陆
func AccountLoginCode(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/account/login/code", func(c *gin.Context) {
		code := c.Query("code")
		id := int64(0)
		uidS := ""
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		ding := ""
		openid := ""
		nameS := ""
		power := int64(-1)
		department := ""
		tokenS := ""

		resp, _ := http.Get("https://api.weixin.qq.com/sns/jscode2session?" +
			"appid=wx08c2e30623d360e9" +
			"&secret=c9875679121476cdbf99f09c944b5122" +
			"&js_code=" + code +
			"&grant_type=authorization_code")
		defer resp.Body.Close()              // 函数结束时关闭Body
		body, _ := ioutil.ReadAll(resp.Body) // 读取Body
		var WX wx
		json.Unmarshal(body, &WX)
		openid = WX.Openid

		if openid != Database.SelectKeyGetFieldS(db, TabName, "openid", "openid='"+openid+"'") {
			c.JSON(208, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       208,
				"message":    "未绑定该openid",
			})
		} else {
			stmt, _, _ := Database.SelectAllData(db, TabName, "id,openid,uid,name,power,department,ding", "openid='"+openid+"'")
			defer stmt.Close()
			if stmt.Next() {
				err := stmt.Scan(&id, &openid, &uid, &nameS, &power, &department, &ding)
				if err != nil {
					log.Fatal(err)
				}
				department = Database.SelectKeyGetFieldS(db, "department_list", "name", "id="+department)
				tokenS = token.CreateToken(strconv.FormatInt(id, 10))
				c.JSON(200, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": department,
					"token":      tokenS,
					"code":       200,
					"message":    "success",
				})
			}
		}
	})
}

//openid绑定uid并登陆
func AccountLoginOpenid(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/account/login/openid", func(c *gin.Context) {
		id := int64(0)
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		openid := c.Query("openid")
		ding := ""
		nameS := ""
		power := int64(-1)
		department := ""
		tokenS := ""

		if openid == Database.SelectKeyGetFieldS(db, TabName, "openid", "openid='"+openid+"'") {
			c.JSON(209, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": department,
				"token":      tokenS,
				"code":       209,
				"message":    "openid已绑定电话号,无法进行二次绑定",
			})
		} else {
			if uid != Database.SelectKeyGetFieldI(db, TabName, "uid", "uid="+uidS) {
				c.JSON(210, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": department,
					"token":      tokenS,
					"code":       210,
					"message":    "该手机号未录入无法进行绑定",
				})
			} else {
				if Database.SelectKeyGetFieldS(db, TabName, "openid", "uid="+uidS) != "" {
					c.JSON(211, gin.H{
						"id":         id,
						"openid":     openid,
						"uid":        uid,
						"ding":       ding,
						"name":       nameS,
						"power":      power,
						"department": department,
						"token":      tokenS,
						"code":       211,
						"message":    "该手机号已经绑定openid,无法二次绑定",
					})
				} else {
					Database.UpdateData(db, TabName, "openid='"+openid+"'", "uid="+uidS)
					stmt, _, _ := Database.SelectAllData(db, TabName, "id,openid,uid,name,power,department,ding", "openid='"+openid+"'")
					defer stmt.Close()
					if stmt.Next() {
						err := stmt.Scan(&id, &openid, &uid, &nameS, &power, &department,&ding)
						if err != nil {
							log.Fatal(err)
						}
						department = Database.SelectKeyGetFieldS(db, "department_list", "name", "id="+department)
						tokenS = token.CreateToken(strconv.FormatInt(id, 10))
						c.JSON(200, gin.H{
							"id":         id,
							"openid":     openid,
							"uid":        uid,
							"ding":       ding,
							"name":       nameS,
							"power":      power,
							"department": department,
							"token":      tokenS,
							"code":       200,
							"message":    "success",
						})
					}
				}
			}
		}
	})
}

//更改个人信息
func AccountChange(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/account/change", func(c *gin.Context) {
		id := int64(0)
		idS := ""
		openid := ""
		uid := int64(0)
		nameS := c.Query("name")
		ding := c.Query("ding")
		departmentName := c.Query("department")
		departmentS := ""
		department := int64(0)
		power := int64(-1)
		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		nameS = ChangeToChinese(nameS)
		departmentName = ChangeToChinese(departmentName)
		if nameS == "" {
			nameS = Database.SelectKeyGetFieldS(db, TabName, "name", "id="+p.UserID)
		}
		if departmentName == "" {
			department = Database.SelectKeyGetFieldI(db, TabName, "department", "id="+p.UserID)
			departmentS = strconv.FormatInt(department, 10)
			departmentName = Database.SelectKeyGetFieldS(db, "department_list", "name", "id="+departmentS)
		} else {
			departmentS = Database.SelectKeyGetFieldS(db, "department_list", "id", "name='"+departmentName+"'")
			department, _ = strconv.ParseInt(departmentS, 10, 64)
		}
		if ding == "" {
			ding = Database.SelectKeyGetFieldS(db, TabName, "ding", "id="+p.UserID)
		}
		fmt.Println("name:"+nameS)
		fmt.Println("department:="+departmentName)
		fmt.Println("ding:"+ding)
		if err != nil {
			c.JSON(201, gin.H{
				"id":         id,
				"openid":     openid,
				"uid":        uid,
				"ding":       ding,
				"name":       nameS,
				"power":      power,
				"department": departmentS,
				"token":      tokenS,
				"code":       201,
				"message":    "token解析错误",
			})
		} else {
			idS = p.UserID
			id, _ := strconv.ParseInt(idS, 10, 64)
			if id != Database.SelectKeyGetFieldI(db, TabName, "id", "id="+idS) {
				c.JSON(200, gin.H{
					"id":         id,
					"openid":     openid,
					"uid":        uid,
					"ding":       ding,
					"name":       nameS,
					"power":      power,
					"department": departmentS,
					"token":      tokenS,
					"code":       204,
					"message":    "账号不存在",
				})
			} else {
				_, _ = Database.UpdateData(db, TabName, "name='"+nameS+"',department="+departmentS+",ding='"+ding+"'", "id="+idS)
				stmt, _, _ := Database.SelectAllData(db, TabName, "*", "id="+idS)
				defer stmt.Close()
				if stmt.Next() {
					err := stmt.Scan(&id, &openid, &uid, &ding, &nameS, &power, &department)
					if err != nil {
						log.Fatal(err)
					}
					tokenS = token.CreateToken(strconv.FormatInt(id, 10))
					departmentS = Database.SelectKeyGetFieldS(db, "department_list", "name", "id="+strconv.FormatInt(department, 10))
					c.JSON(200, gin.H{
						"id":         id,
						"openid":     openid,
						"uid":        uid,
						"ding":       ding,
						"name":       nameS,
						"power":      power,
						"department": departmentS,
						"token":      tokenS,
						"code":       200,
						"message":    "success",
					})
				}
			}
		}
	})
}

//获取订阅情况
func SubscribeGet(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/subscribe/get", func(c *gin.Context) {
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		if uid == Database.SelectKeyGetFieldI(db, "subscribe_list", "uid", "uid="+uidS) {
			if Database.SelectKeyGetFieldI(db, "subscribe_list", "subscribe", "uid="+uidS) == 1 {
				c.JSON(200, gin.H{
					"subscribe": 1,
					"code":      200,
					"message":   "已订阅",
				})
			} else {
				c.JSON(200, gin.H{
					"subscribe": 0,
					"code":      200,
					"message":   "未订阅",
				})
			}
		} else {
			c.JSON(212, gin.H{
				"subscribe": 0,
				"code":      212,
				"message":   "查无此人",
			})
		}
	})
}

//更改为已订阅
func SubscribeChange(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/subscribe/change", func(c *gin.Context) {
		uidS := c.Query("uid")
		uid, _ := strconv.ParseInt(uidS, 10, 64)
		if uid == Database.SelectKeyGetFieldI(db, "subscribe_list", "uid", "uid="+uidS) {
			Database.UpdateData(db, "subscribe_list", "subscribe=1", "uid="+uidS)
			c.JSON(200, gin.H{
				"subscribe": 1,
				"code":      200,
				"message":   "success",
			})
		} else {
			c.JSON(212, gin.H{
				"subscribe": 0,
				"code":      212,
				"message":   "查无此人",
			})
		}
	})
}

//用户信息管理系统
func AccountSystem(r *gin.Engine, db *sql.DB, TabName string) {
	AccountAdd(r, TabName, db)
	AccountLoginCode(r, TabName, db)
	AccountLoginOpenid(r, TabName, db)
	AccountChange(r, TabName, db)
	SubscribeGet(r, TabName, db)
	SubscribeChange(r, TabName, db)
}

//部门系统
//添加部门
func DepartmentAdd(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/department/add", func(c *gin.Context) {
		id := int64(0)
		idS := "0"
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"id":      id,
				"name":    nameS,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			idS = p.UserID
			id, _ = strconv.ParseInt(idS, 10, 64)
			if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+idS) != 0 {
				c.JSON(202, gin.H{
					"id":      id,
					"name":    nameS,
					"code":    202,
					"message": "非超管无权限",
				})
			} else {
				Database.InsertData(db, TabName, "0,'"+nameS+"'")
				stmt, _, _ := Database.SelectLastData(db, "*", TabName)
				defer stmt.Close()
				if stmt.Next() {
					err := stmt.Scan(&id, &nameS)
					if err != nil {
						log.Fatal(err)
					}
					c.JSON(200, gin.H{
						"id":      id,
						"name":    nameS,
						"code":    200,
						"message": "success",
					})
				}
			}
		}

	})
}

//删除部门
func DepartmentDelete(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/department/delete", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		nameS := ""

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"id":      id,
				"name":    nameS,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			if id != Database.SelectKeyGetFieldI(db, TabName, "id", "id="+idS) {
				c.JSON(204, gin.H{
					"id":      id,
					"name":    nameS,
					"code":    204,
					"message": "该部门不存在",
				})
			} else {
				if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID) != 0 {
					c.JSON(202, gin.H{
						"id":      id,
						"name":    nameS,
						"code":    202,
						"message": "非超管无权限",
					})
				} else {
					memberList := Database.SelectKeyGetFieldsI(db, "account_list", "*", "department="+idS)
					contractList := Database.SelectKeyGetFieldsI(db, "contract_list", "*", "department="+idS)
					for n := range memberList {
						Database.UpdateData(db, TabName, "department=0", "id="+strconv.FormatInt(memberList[n], 10))
					}
					for n := range memberList {
						Database.UpdateData(db, TabName, "department=0", "id="+strconv.FormatInt(contractList[n], 10))
					}
					Database.DeleteData(db, TabName, "id="+idS)
					c.JSON(200, gin.H{
						"id":      id,
						"name":    nameS,
						"code":    200,
						"message": "success",
					})
				}
			}
		}
	})
}

//更改部门名称
func DepartmentChangeName(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/department/change/name", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if strings.Contains(nameS, "\\u") {
			sUnicodev := strings.Split(nameS, "\\u")
			var context string
			for _, v := range sUnicodev {
				if len(v) < 1 {
					continue
				}
				temp, err := strconv.ParseInt(v, 16, 32)
				if err != nil {
					panic(err)
				}
				context += fmt.Sprintf("%c", temp)
			}
			nameS = context
		}
		if err != nil {
			c.JSON(201, gin.H{
				"id":      id,
				"name":    nameS,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			if id != Database.SelectKeyGetFieldI(db, TabName, "id", "id="+idS) {
				c.JSON(204, gin.H{
					"id":      id,
					"name":    nameS,
					"code":    204,
					"message": "该部门不存在",
				})
			} else {
				if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID) != 0 {
					c.JSON(202, gin.H{
						"id":      id,
						"name":    nameS,
						"code":    202,
						"message": "非超管无权限",
					})
				} else {
					Database.UpdateData(db, TabName, "name='"+nameS+"'", "id="+idS)
					c.JSON(200, gin.H{
						"id":      id,
						"name":    nameS,
						"code":    200,
						"message": "success",
					})
				}
			}
		}
	})
}

//查看部门列表
func DepartmentList(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/department/list", func(c *gin.Context) {
		tokenS := c.Query("token")
		_, err := token.CheckToken(tokenS)
		idList := make([]int64, 0)
		nameList := make([]string, 0)
		if err != nil {
			c.JSON(201, gin.H{
				"idList":   idList,
				"nameList": nameList,
				"code":     201,
				"message":  "token解析错误",
			})
		} else {
			stmt, _, _ := Database.SelectAllData(db, TabName, "*", "")
			defer stmt.Close()
			for stmt.Next() {
				var id int64
				var name string
				err := stmt.Scan(&id, &name)
				if err != nil {
					log.Fatal(err)
				}
				idList = append(idList, id)
				nameList = append(nameList, name)
			}
			c.JSON(200, gin.H{
				"idList":   idList,
				"nameList": nameList,
				"code":     200,
				"message":  "success",
			})
		}
	})
}

//部门信息管理系统
func DepartmentSystem(r *gin.Engine, db *sql.DB, TabName string) {
	DepartmentAdd(r, TabName, db)
	DepartmentDelete(r, TabName, db)
	DepartmentChangeName(r, TabName, db)
	DepartmentList(r, TabName, db)
}

//合同系统
//合同类型列表
func ContractTypeList(r *gin.Engine) {
	r.GET("/contract/type", func(c *gin.Context) {
		list:=make([]string,0)
		list=append(list,"工程项目合同")
		list=append(list,"物资、设备、服务采购合同")
		list=append(list,"招标代理合同")
		list=append(list,"技术开发类")
		list=append(list,"租赁合同")
		list=append(list,"捐赠合同")
		list=append(list,"维修承揽合同")
		list=append(list,"其他合同")
		c.JSON(200, gin.H{
			"list":list,
		})
	})
}

//合同类型列表
func ContractTimeList(r *gin.Engine) {
	r.GET("/contract/time", func(c *gin.Context) {
		list:=make([]string,0)
		list=append(list,"缴纳租金时间")
		list=append(list,"缴存投标保证金时间")
		list=append(list,"退回投标保证金时间")
		list=append(list,"缴存履约保证金时间")
		list=append(list,"退回履约保证金时间")
		list=append(list,"缴存质量保证金时间")
		list=append(list,"退回质量保证金时间")
		list=append(list,"扣除违约金时间")
		list=append(list,"验收时间")
		list=append(list,"结算审计时间")
		list=append(list,"第一次付款时间")
		list=append(list,"第二次付款时间")
		c.JSON(200, gin.H{
			"list":list,
		})
	})
}

//添加合同
func ContractAdd(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/add/all", func(c *gin.Context) {
		id := int64(0)
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)
		notesS := c.Query("notes")
		notesS = ChangeToChinese(notesS)
		typeS := c.Query("type")
		uploader := int64(0)
		department := int64(0)
		begin := toTimestamp(c.Query("begin"))
		deadline := toTimestamp(c.Query("deadline"))
		remindTime := c.Query("remindTime")
		remindTime = ChangeToChinese(remindTime)
		RT := make([]remind, 0)
		json.Unmarshal([]byte(remindTime), &RT)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)

		uploader, _ = strconv.ParseInt(p.UserID, 10, 64)
		if err != nil {
			c.JSON(201, gin.H{
				"id":         id,
				"name":       nameS,
				"notes":      notesS,
				"type":       typeS,
				"uploader":   uploader,
				"department": department,
				"begin":      begin,
				"deadline":   deadline,
				"remindTime": RT,
				"code":       201,
				"message":    "token解析错误",
			})
		} else {
			department = Database.SelectKeyGetFieldI(db, "account_list", "department", "id="+p.UserID)
			if department == 0 {
				c.JSON(207, gin.H{
					"id":         id,
					"name":       nameS,
					"notes":      notesS,
					"type":       typeS,
					"uploader":   uploader,
					"department": department,
					"begin":      begin,
					"deadline":   deadline,
					"remindTime": RT,
					"code":       207,
					"message":    "该用户未加入部门",
				})
			} else {
				Database.InsertData(db, TabName, "0,'"+nameS+"','"+notesS+"','"+typeS+"',"+strconv.FormatInt(uploader, 10)+","+strconv.FormatInt(department, 10)+",''")
				stmt, _, _ := Database.SelectLastData(db, "id", TabName)
				defer stmt.Close()
				if stmt.Next() {
					err := stmt.Scan(&id)
					if err != nil {
						log.Fatal(err)
					}
					Database.InsertData(db, "time_list", "0,'begin',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(begin, 10))
					Database.InsertData(db, "time_list", "0,'deadline',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(deadline, 10))
					for x := range RT {
						timeStampS := strconv.FormatInt(toTimestamp(RT[x].Time), 10)
						Database.InsertData(db, "time_list", "0,'"+RT[x].Name+"',"+strconv.FormatInt(id, 10)+","+timeStampS)
					}
					c.JSON(200, gin.H{
						"id":         id,
						"name":       nameS,
						"notes":      notesS,
						"type":       typeS,
						"uploader":   uploader,
						"department": department,
						"begin":      begin,
						"deadline":   deadline,
						"remindTime": RT,
						"code":       200,
						"message":    "success",
					})
				}
			}
		}
	})
}

//添加合同PDF
func ContractAddPDF(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.POST("/contract/add/pdf/:id", func(c *gin.Context) {
		idS := c.Param("id")
		id, _ := strconv.ParseInt(idS, 10, 64)

		pdf, errpdf := c.FormFile("pdf")
		pdfType := ""
		if errpdf != nil {
			c.JSON(206, gin.H{
				"id":       id,
				"contract": id,
				"code":     206,
				"message":  "获取合同失败",
			})
		} else {
			pdfType = strings.Split(pdf.Filename, ".")[len(strings.Split(pdf.Filename, "."))-1]
			Database.UpdateData(db, TabName, "pdfType='"+pdfType+"'", "id="+idS)
			if err := c.SaveUploadedFile(pdf, path+"/PDF/"+strconv.FormatInt(id, 10)+"."+pdfType); err != nil {
				c.JSON(206, gin.H{
					"id":       id,
					"contract": id,
					"code":     206,
					"message":  "保存合同失败",
				})
			} else {
				c.JSON(200, gin.H{
					"id":       id,
					"contract": id,
					"code":     200,
					"message":  "success",
				})
			}
		}
	})
}

//合同添加图片
func ContractAddImg(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.POST("/contract/add/img/:contract", func(c *gin.Context) {
		id := int64(0)
		contractS := c.Param("contract")
		contract, _ := strconv.ParseInt(contractS, 10, 64)

		img, errimg := c.FormFile("img")
		imgType := ""
		if errimg != nil {
			c.JSON(205, gin.H{
				"id":       id,
				"contract": contract,
				"code":     205,
				"message":  "获取图片失败",
			})
		} else {
			imgType = strings.Split(img.Filename, ".")[len(strings.Split(img.Filename, "."))-1]
			Database.InsertData(db, "img_list", "0,"+contractS+",'"+imgType+"'")
			stmt, _, _ := Database.SelectLastData(db, "id", "img_list")
			defer stmt.Close()
			if stmt.Next() {
				err := stmt.Scan(&id)
				if err != nil {
					log.Fatal(err)
				}
				if err := c.SaveUploadedFile(img, path+"/IMG/"+strconv.FormatInt(id, 10)+"."+imgType); err != nil {
					c.JSON(205, gin.H{
						"id":       id,
						"contract": contract,
						"code":     205,
						"message":  "保存图片失败",
					})
				} else {
					c.JSON(200, gin.H{
						"id":       id,
						"contract": contract,
						"code":     200,
						"message":  "success",
					})
				}
			}
		}
	})
}

//搜索部门合同
func ContractFindDepartment(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/find/department", func(c *gin.Context) {
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)
		list := make([]contract, 0)
		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"list":    list,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			power := Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID)
			department := Database.SelectKeyGetFieldI(db, "account_list", "department", "id="+p.UserID)
			stmt, _, _ := Database.SelectAllData(db, TabName, "id,name,notes,department,uploader,type", "name like '%"+nameS+"%'")
			defer stmt.Close()
			for stmt.Next() {
				var id, d, up int64
				var name, notes, deadline, begin, typeS string
				err := stmt.Scan(&id, &name, &notes, &d, &up, &typeS)
				if err != nil {
					log.Fatal(err)
				}
				if power == 0 || department == d {
					uploader := Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+strconv.FormatInt(up, 10))
					stmt, _, _ := Database.SelectAllData(db, "time_list", "name,timestamp", "contract="+strconv.FormatInt(id, 10))
					RT := make([]remind, 0)
					for stmt.Next() {
						name := ""
						time := int64(0)
						err := stmt.Scan(&name, &time)
						if err != nil {
							log.Fatal(err)
						}
						timeS := toTimeS(time)
						if name == "deadline" {
							deadline = timeS
						} else if name == "begin" {
							begin = timeS
						} else {
							RT = append(RT, remind{
								Name: name,
								Time: timeS,
							})
						}
					}
					stmt.Close()
					list = append(list, contract{
						ID:       id,
						Name:     name,
						Notes:    notes,
						Type:     typeS,
						Uploader: uploader,
						Begin:    begin,
						Deadline: deadline,
						Remind:   RT,
					})
				}
			}
			c.JSON(200, gin.H{
				"list":    list,
				"code":    200,
				"message": "success",
			})
		}
	})
}

//搜索私人合同
func ContractFindPrivate(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/find/private", func(c *gin.Context) {
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)
		list := make([]contract, 0)
		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"list":    list,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			power := Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID)
			stmt, _, _ := Database.SelectAllData(db, TabName, "id,name,notes,department,uploader,type", "name like '%"+nameS+"%'")
			defer stmt.Close()
			for stmt.Next() {
				var id, d, up int64
				var name, notes, deadline, begin, typeS string
				err := stmt.Scan(&id, &name, &notes, &d, &up, &typeS)
				if err != nil {
					log.Fatal(err)
				}
				upS := strconv.FormatInt(up, 10)
				if power == 0 || upS == p.UserID {
					uploader := Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+strconv.FormatInt(up, 10))
					stmt, _, _ := Database.SelectAllData(db, "time_list", "name,timestamp", "contract="+strconv.FormatInt(id, 10))
					RT := make([]remind, 0)
					for stmt.Next() {
						name := ""
						time := int64(0)
						err := stmt.Scan(&name, &time)
						if err != nil {
							log.Fatal(err)
						}
						timeS := toTimeS(time)
						if name == "deadline" {
							deadline = timeS
						} else if name == "begin" {
							begin = timeS
						} else {
							RT = append(RT, remind{
								Name: name,
								Time: timeS,
							})
						}
					}
					stmt.Close()
					list = append(list, contract{
						ID:       id,
						Name:     name,
						Notes:    notes,
						Type:     typeS,
						Uploader: uploader,
						Begin:    begin,
						Deadline: deadline,
						Remind:   RT,
					})
				}
			}
			c.JSON(200, gin.H{
				"list":    list,
				"code":    200,
				"message": "success",
			})
		}
	})
}

//个人最近合同
func ContractLately(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/lately", func(c *gin.Context) {
		list := make([]contract, 0)
		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"list":    list,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			stmt, _, _ := Database.SelectAllData(db, TabName, "id,name,notes,department,type", "uploader="+p.UserID)
			defer stmt.Close()
			for stmt.Next() {
				var id, d int64
				var name, notes, deadline, begin, typeS string
				err := stmt.Scan(&id, &name, &notes, &d, &typeS)
				if err != nil {
					log.Fatal(err)
				}
				uploader := Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+p.UserID)
				rows, _, _ := Database.SelectAllData(db, "time_list", "name,timestamp", "contract="+strconv.FormatInt(id, 10))
				defer rows.Close()
				RT := make([]remind, 0)
				for rows.Next() {
					name1 := ""
					time1 := int64(0)
					err := rows.Scan(&name1, &time1)
					if err != nil {
						log.Fatal(err)
					}
					timeS := toTimeS(time1)
					if name1 == "deadline" {
						deadline = timeS
					} else if name1 == "begin" {
						begin = timeS
					} else {
						RT = append(RT, remind{
							Name: name1,
							Time: timeS,
						})
					}
				}
				list = append(list, contract{
					ID:       id,
					Name:     name,
					Notes:    notes,
					Type:     typeS,
					Uploader: uploader,
					Begin:    begin,
					Deadline: deadline,
					Remind:   RT,
				})
			}
			if len(list) > 5 {
				list = list[0:5]
			}
			for x := len(list) - 1; x >= 0; x-- {
				list = append(list, list[x])
			}
			list = list[len(list)/2:]
			c.JSON(200, gin.H{
				"list":    list,
				"code":    200,
				"message": "success",
			})
		}
	})
}

//合同编辑
func ContractEdit(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/edit", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		nameS := c.Query("name")
		nameS = ChangeToChinese(nameS)
		notesS := c.Query("notes")
		notesS = ChangeToChinese(notesS)
		typeS := c.Query("type")
		uploader := int64(0)
		department := int64(0)
		beginS := c.Query("begin")
		begin := toTimestamp(beginS)
		deadlineS := c.Query("deadline")
		deadline := toTimestamp(deadlineS)
		remindTime := c.Query("remindTime")
		remindTime = ChangeToChinese(remindTime)
		RT := make([]remind, 0)
		json.Unmarshal([]byte(remindTime), &RT)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)

		uploader, _ = strconv.ParseInt(p.UserID, 10, 64)

		if err != nil {
			c.JSON(201, gin.H{
				"id":         id,
				"name":       nameS,
				"notes":      notesS,
				"type":       typeS,
				"uploader":   uploader,
				"department": department,
				"begin":      beginS,
				"deadline":   deadlineS,
				"remindTime": RT,
				"code":       201,
				"message":    "token解析错误",
			})
		} else {
			if id != Database.SelectKeyGetFieldI(db, TabName, "id", "id="+idS) {
				c.JSON(204, gin.H{
					"id":         id,
					"name":       nameS,
					"notes":      notesS,
					"type":       typeS,
					"uploader":   uploader,
					"department": department,
					"begin":      beginS,
					"deadline":   deadlineS,
					"remindTime": RT,
					"code":       204,
					"message":    "合同不存在",
				})
			} else {
				if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID) != 0 && Database.SelectKeyGetFieldI(db, TabName, "uploader", "id="+idS) != uploader {
					c.JSON(202, gin.H{
						"id":         id,
						"name":       nameS,
						"notes":      notesS,
						"type":       typeS,
						"uploader":   uploader,
						"department": department,
						"begin":      beginS,
						"deadline":   deadlineS,
						"remindTime": RT,
						"code":       202,
						"message":    "非超管也非合同上传者,无权限编辑",
					})
				} else {
					if nameS == "" {
						nameS = Database.SelectKeyGetFieldS(db, TabName, "name", "id="+idS)
					}
					if notesS == "" {
						notesS = Database.SelectKeyGetFieldS(db, TabName, "notes", "id="+idS)
					}
					if typeS == "点击选择合同类型" {
						typeS = Database.SelectKeyGetFieldS(db, TabName, "type", "id="+idS)
					}
					if begin <= 0 {
						begin = Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+idS+" and name='begin'")
						beginS = toTimeS(begin)
					}
					if deadline <= 0 {
						deadline = Database.SelectKeyGetFieldI(db, "time_list", "timestamp", "contract="+idS+" and name='deadline'")
						deadlineS = toTimeS(deadline)
					}
					Database.UpdateData(db, TabName, "name='"+nameS+"',notes='"+notesS+"',type='"+typeS+"'", "id="+idS)
					if len(RT) != 0 {
						Database.DeleteData(db, "time_list", "contract="+idS)
						Database.InsertData(db, "time_list", "0,'begin',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(begin, 10))
						Database.InsertData(db, "time_list", "0,'deadline',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(deadline, 10))
					}else{
						Database.DeleteData(db, "time_list", "contract="+idS+" and name='begin'")
						Database.DeleteData(db, "time_list", "contract="+idS+" and name='deadline'")
						Database.InsertData(db, "time_list", "0,'begin',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(begin, 10))
						Database.InsertData(db, "time_list", "0,'deadline',"+strconv.FormatInt(id, 10)+","+strconv.FormatInt(deadline, 10))
					}
					for x := range RT {
						timeStampS := strconv.FormatInt(toTimestamp(RT[x].Time), 10)
						Database.InsertData(db, "time_list", "0,'"+RT[x].Name+"',"+strconv.FormatInt(id, 10)+","+timeStampS)
					}
					c.JSON(200, gin.H{
						"id":         id,
						"name":       nameS,
						"notes":      notesS,
						"type":       typeS,
						"uploader":   uploader,
						"department": department,
						"begin":      beginS,
						"deadline":   deadlineS,
						"remindTime": RT,
						"code":       200,
						"message":    "success",
					})
				}
			}
		}
	})
}

//合同展示
func ContractShowAll(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/show/all", func(c *gin.Context) {
		uid := int64(0)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		nameS := ""
		notesS := ""
		typeS := ""
		uploader := int64(0)
		uploaderS := ""
		department := int64(0)
		begin := ""
		deadline := ""
		pdf := ""
		RT := make([]remind, 0)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(200, gin.H{
				"uid":        uid,
				"id":         id,
				"name":       nameS,
				"notes":      notesS,
				"type":       typeS,
				"uploader":   uploaderS,
				"department": department,
				"begin":      begin,
				"deadline":   deadline,
				"remindTime": RT,
				"pdf":        pdf,
				"code":       201,
				"message":    "token解析错误",
			})
		} else {
			power := Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID)
			departmentU := Database.SelectKeyGetFieldI(db, "account_list", "department", "id="+p.UserID)

			stmt, _, _ := Database.SelectAllData(db, TabName, "id,name,notes,uploader,department,pdfType,type", "id="+idS)
			defer stmt.Close()
			if stmt.Next() {
				err := stmt.Scan(&id, &nameS, &notesS, &uploader, &department, &pdf, &typeS)
				if err != nil {
					log.Fatal(err)
				}
				if power == 0 || departmentU == department {
					stmt, _, _ := Database.SelectAllData(db, "time_list", "name,timestamp", "contract="+idS)
					for stmt.Next() {
						name := ""
						time := int64(0)
						err := stmt.Scan(&name, &time)
						if err != nil {
							log.Fatal(err)
						}
						timeS := toTimeS(time)
						if name == "deadline" {
							deadline = timeS
						} else if name == "begin" {
							begin = timeS
						} else {
							var R remind
							R.Name = name
							R.Time = timeS
							RT = append(RT, remind{
								Name: name,
								Time: timeS,
							})
						}
					}
					stmt.Close()
					uploaderS = Database.SelectKeyGetFieldS(db, "account_list", "name", "id="+strconv.FormatInt(uploader, 10))
					uid = Database.SelectKeyGetFieldI(db, "account_list", "uid", "id="+strconv.FormatInt(uploader, 10))
					c.JSON(200, gin.H{
						"uid":        uid,
						"id":         id,
						"name":       nameS,
						"notes":      notesS,
						"type":       typeS,
						"uploader":   uploaderS,
						"department": department,
						"begin":      begin,
						"deadline":   deadline,
						"remindTime": RT,
						"pdf":        pdf,
						"code":       200,
						"message":    "success",
					})
				} else {
					c.JSON(200, gin.H{
						"uid":        uid,
						"id":         id,
						"name":       "",
						"notes":      "",
						"type":       "",
						"uploader":   0,
						"department": 0,
						"begin":      "",
						"deadline":   "",
						"remindTime": nil,
						"pdf":        "",
						"code":       202,
						"message":    "无权查看",
					})
				}
			}
		}
	})
}

//合同IMG列表
func ContractShowList(r *gin.Engine, TabName string, db *sql.DB) {
	r.GET("/contract/show/list", func(c *gin.Context) {
		contractS := c.Query("contract")
		stmt, _, _ := Database.SelectAllData(db, TabName, "id", "contract='"+contractS+"'")
		defer stmt.Close()
		idList := make([]int64, 0)
		urlList := make([]string, 0)
		for stmt.Next() {
			var id int64
			err := stmt.Scan(&id)
			if err != nil {
				log.Fatal(err)
			}
			s := "http://47.108.217.244:2975/contract/show/img/" + strconv.FormatInt(id, 10)
			urlList = append(urlList, s)
			idList = append(idList, id)
		}
		c.JSON(200, gin.H{
			"id_list":  idList,
			"url_list": urlList,
			"code":     200,
			"message":  "success",
		})
	})
}

//合同IMG展示
func ContractShowIMG(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.GET("/contract/show/img/:id", func(c *gin.Context) {
		idS := c.Param("id")
		imgType := ""
		stmt, _, _ := Database.SelectAllData(db, TabName, "imgType", "id="+idS)
		defer stmt.Close()
		if stmt.Next() {
			err := stmt.Scan(&imgType)
			if err != nil {
				log.Fatal(err)
			}
			file, _ := ioutil.ReadFile(path + "/IMG/" + idS + "." + imgType)
			c.Writer.WriteString(string(file))
		}
	})
}

//合同PDF展示
func ContractShowPDF(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.GET("/contract/show/pdf/:id", func(c *gin.Context) {
		idS := c.Param("id")
		department := int64(0)
		pdfType := ""
		stmt, _, _ := Database.SelectAllData(db, TabName, "department,pdfType", "id="+idS)
		defer stmt.Close()
		if stmt.Next() {
			err := stmt.Scan(&department, &pdfType)
			if err != nil {
				log.Fatal(err)
			}
			file, _ := ioutil.ReadFile(path + "/PDF/" + idS + "." + pdfType)
			c.Writer.WriteString(string(file))
		}
	})
}

//合同图片删除
func ContractDeleteIMG(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.GET("/contract/delete/img", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"id":      id,
				"code":    201,
				"message": "token解析错误",
			})
		} else {
			if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID) != 0 && Database.SelectKeyGetFieldS(db, TabName, "uploader", "id="+idS) != p.UserID {
				c.JSON(202, gin.H{
					"id":      id,
					"code":    202,
					"message": "无权删除",
				})
			} else {
				imgType := Database.SelectKeyGetFieldS(db, "img_list", "imgType", "id="+idS+"")
				os.Remove(path + "/IMG/" + idS + "." + imgType)
				c.JSON(200, gin.H{
					"id":      id,
					"code":    200,
					"message": "success",
				})
			}
		}
	})
}

//合同删除
func ContractDelete(r *gin.Engine, TabName string, db *sql.DB, path string) {
	r.GET("/contract/delete/all", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		nameS := ""
		notesS := ""
		uploader := int64(0)
		department := int64(0)

		tokenS := c.Query("token")
		p, err := token.CheckToken(tokenS)
		if err != nil {
			c.JSON(201, gin.H{
				"id":         id,
				"name":       nameS,
				"notes":      notesS,
				"uploader":   uploader,
				"department": department,
				"code":       201,
				"message":    "token解析错误",
			})
		} else {
			if Database.SelectKeyGetFieldI(db, TabName, "id", "id="+idS) != id {
				c.JSON(204, gin.H{
					"id":         id,
					"name":       nameS,
					"notes":      notesS,
					"uploader":   uploader,
					"department": department,
					"code":       204,
					"message":    "该合同不存在",
				})
			} else {
				if Database.SelectKeyGetFieldI(db, "account_list", "power", "id="+p.UserID) != 0 && Database.SelectKeyGetFieldS(db, TabName, "uploader", "id="+idS) != p.UserID {
					c.JSON(202, gin.H{
						"id":         id,
						"name":       nameS,
						"notes":      notesS,
						"uploader":   uploader,
						"department": department,
						"code":       202,
						"message":    "无权删除",
					})
				} else {
					stmt, _, _ := Database.SelectAllData(db, TabName, "pdfType", "id="+idS)
					defer stmt.Close()
					if stmt.Next() {
						var pdfType string
						err := stmt.Scan(&pdfType)
						if err != nil {
							log.Fatal(err)
						}
						a, _, _ := Database.SelectAllData(db, "img_list", "id,imgType", "contract="+idS+"")
						defer a.Close()
						for a.Next() {
							var tmp int64
							var s string
							err := a.Scan(&tmp, &s)
							if err != nil {
								log.Fatal(err)
							}
							os.Remove(path + "/IMG/" + strconv.FormatInt(tmp, 10) + "." + s)
						}
						os.Remove(path + "/PDF/" + idS + "." + pdfType)
						Database.DeleteData(db, TabName, "id="+idS)
						stmt, _, _ := Database.SelectAllData(db, "time_list", "id", "contract="+idS)
						defer stmt.Close()
						tids := make([]int64, 0)
						for stmt.Next() {
							tid := int64(0)
							err := stmt.Scan(&tid)
							if err != nil {
								log.Fatal(err)
							}
							tids = append(tids, tid)
						}
						for x := range tids {
							Database.DeleteData(db, "time_list", "id="+strconv.FormatInt(tids[x], 10))
						}
						c.JSON(200, gin.H{
							"id":         id,
							"name":       nameS,
							"notes":      notesS,
							"uploader":   uploader,
							"department": department,
							"code":       200,
							"message":    "success",
						})
					}
				}
			}
		}
	})
}

//合同信息管理系统
func ContractSystem(r *gin.Engine, db *sql.DB, TabName string) {
	path := "/root/file"
	ContractTypeList(r)
	ContractTimeList(r)
	ContractAdd(r, TabName, db)
	ContractAddPDF(r, TabName, db, path)
	ContractAddImg(r, TabName, db, path)
	ContractFindDepartment(r, TabName, db)
	ContractFindPrivate(r, TabName, db)
	ContractLately(r, TabName, db)
	ContractEdit(r, TabName, db)
	ContractShowAll(r, TabName, db)
	ContractShowList(r, "img_list", db)
	ContractShowIMG(r, "img_list", db, path)
	ContractShowPDF(r, TabName, db, path)
	ContractDeleteIMG(r, TabName, db, path)
	ContractDelete(r, TabName, db, path)
}
func main() {
	r := gin.Default()
	_, _, db := CreateList()
	BackstageSystem(r, db, "account_list")
	AccountSystem(r, db, "account_list")
	DepartmentSystem(r, db, "department_list")
	ContractSystem(r, db, "contract_list")
	r.Run(":2975")
}
