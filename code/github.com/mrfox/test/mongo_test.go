/**
 * @Author: mrfox
 * @Description:
 * @File:  mongo_test
 * @Version: 1.0.0
 * @Date: 2020/3/14 9:39 下午
 */
package test

import (
	"context"
	"github.com/prometheus/common/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type person struct {
	Name string `bson:"p_name"` //姓名
	Age  int    `bson:"p_age"`  //年龄
}

type FinaByPName struct {
	Name string `bson:"p_name"`//根据名字查询
}

//大于
type IntBeforeCond struct {
	Before int64 `bson:"$gt"`
}

//年龄大于x的 {"p_age":{"$gt":x}}
type DeleteByCond struct {
	beforeCond  IntBeforeCond `bson:"p_age"`
}

func Test_mongo(t *testing.T) {
	var (
		collection *mongo.Collection
		//插入结果
		insertOne *mongo.InsertOneResult
		//多条存放数组
		persons []interface{}
		//插入多个结果
		insertMany *mongo.InsertManyResult

		//查询条件
		cond *FinaByPName
		//游标
		cursor *mongo.Cursor

		//删除符合条件的
		deleteMany *mongo.DeleteResult
	)

	//指定连接配置
	//clientOptions := options.Client().ApplyURI("mongodb://cron:cron@10.211.55.4:27017/cron").
	clientOptions := options.Client().ApplyURI("mongodb://10.211.55.4:27017/cron").
		//连接超时时间
		SetConnectTimeout(time.Second * 5).
		SetAuth(options.Credential{Username:"cron",Password:"cron"})
	//连接
	client, err := mongo.Connect(context.TODO(), clientOptions)

	log.Info("[mongo] connect is success")

	if err != nil {
		log.Fatal(err)
	}
	//指定数据库和集合
	collection = client.Database("cron").Collection("mrfox_cron")
	//插入单条记录
	insertOne, err = collection.InsertOne(context.TODO(), person{Name: "小黑", Age: 12})
	if err != nil {
		log.Fatal(err)
	}else {
		log.Infof("[mongo]insert objectId is [%v]",insertOne.InsertedID)
	}

	//初始化数组
	persons = make([]interface{},2)
	persons[0] = person{Name:"小花",Age:12}
	persons[1] = person{Name:"小红",Age:13}

	//插入多个
	if insertMany ,err = collection.InsertMany(context.TODO(),persons);err != nil{
		log.Errorf("[mongo]insert many is error:[%v]",err)
		return
	}

	for _,objId := range insertMany.InsertedIDs{
		log.Infof("[mongo]objId is [%v]",objId)
	}

	log.Info("---------------")

	//根据名字查询出2条
	cond = &FinaByPName{Name:"小红"}
	//从第N页
	skip := int64(0)
	//查询几条记录
	limit := int64(2)

	if cursor, err = collection.Find(context.TODO(), cond, &options.FindOptions{Skip: &skip, Limit: &limit});err != nil{
		log.Errorf("[mongo] find is err:[%v]",err)
		return
	}

	//最后释放游标
	defer cursor.Close(context.TODO())
	defer client.Disconnect(context.TODO())

	//遍历游标
	for cursor.Next(context.TODO()){
		personTmp := &person{}
		//反序列化
		cursor.Decode(personTmp)
		log.Info(personTmp)
	}


	log.Info("------------删除----------------")
	//删除年龄大于12的
	//第一种
	//collection.DeleteMany(context.TODO(),
	//	bson.M{
	//	"p_age":bson.M{"$gt":11},
	//	})


	//第二种
	var delCond *DeleteByCond
	delCond  = &DeleteByCond{beforeCond:IntBeforeCond{Before:12}}
	if deleteMany, err = collection.DeleteMany(context.TODO(), delCond);err != nil{
		log.Errorf("[mongo]delete mongo doc is error:[%v]",err)
		return
	}
	log.Infof("[mongo]delete mongo doc is success,count:[%v]",deleteMany.DeletedCount)

}