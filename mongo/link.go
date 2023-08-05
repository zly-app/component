package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Client     = mongo.Client
	Database   = mongo.Database
	Collection = mongo.Collection
	Cursor     = mongo.Cursor

	SessionContext = mongo.SessionContext
	Session = mongo.Session
)

var (
	ErrUnacknowledgedWrite = mongo.ErrUnacknowledgedWrite
	ErrClientDisconnected  = mongo.ErrClientDisconnected
	ErrNilDocument         = mongo.ErrNilDocument
	ErrNilValue            = mongo.ErrNilValue
	ErrEmptySlice          = mongo.ErrEmptySlice

	// 来自服务器的更改流通知不包含恢复令牌
	ErrMissingResumeToken = mongo.ErrMissingResumeToken
	// 更改流的底层游标为nil
	ErrNilCursor = mongo.ErrNilCursor

	// 使用具有非数字或字符串值的键文档创建索引
	ErrInvalidIndexValue = mongo.ErrInvalidIndexValue
	// 创建索引的名称不是字符串
	ErrNonStringIndexName = mongo.ErrNonStringIndexName
	// 多个索引将被删除
	ErrMultipleIndexDrop = mongo.ErrMultipleIndexDrop

	// 该客户端没有创建会话
	ErrWrongClient = mongo.ErrWrongClient
)

type (
	BsonD = bson.D
	BsonE = bson.E
	BsonM = bson.M
	BsonA = bson.A
)
