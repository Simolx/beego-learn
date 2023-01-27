package demo

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	//    orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/beego?charset=utf8")
	orm.RegisterDataBase("default", "mysql", "sophonuser:password@tcp(localhost:3306)/metastore_sophon1?charset=utf8")
}

func TestOrm(t *testing.T) {
	fmt.Println("start dev")
	o := orm.NewOrm()
	cmd := o.Raw("show tables")
	res, err := cmd.Exec()
	if err != nil {
		fmt.Printf("cmd: %v run error %v", cmd, err)
		return
	}
	fmt.Printf("exec result %#v", res)
	//    insertCmd := o.Raw("insert into devtable (bid, kmc_key, kmc) values (?, ?, ?)", 2, "abc", "YWJjCg==")
	//    runCmd(insertCmd)
	//    insertCmd2 := o.Raw(`insert into devtable (bid, kmc_key, kmc) values (?, ?, ?)`, 3, "abcd", "YWJjZAo=")
	//    runCmd(insertCmd2)
	//    updateCmd := o.Raw("update devtable set kmc_key=?, kmc=? where bid=?", "abcde", "YWJjZGUK", 2)
	//    runCmd(updateCmd)
	updateCmd2 := o.Raw(`update devtable set kmc_key=?, kmc=? where bid=?`, "abcdefg", "YWJjZGVmZwo=", 3)
	runCmd(updateCmd2)
}

func runCmd(cmd orm.RawSeter) error {
	fmt.Printf("cmd: %v\n", cmd)
	result, err := cmd.Exec()
	if err != nil {
		fmt.Printf("exec cmd error, %v\n", err)
		return err
	}
	rl, _ := result.LastInsertId()
	rr, _ := result.RowsAffected()
	fmt.Printf("exec command result %d, %d\n", rl, rr)
	return nil
}
