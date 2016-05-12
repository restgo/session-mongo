package session_mongo

import (
	"encoding/json"
	"github.com/restgo/jsonhelper"
	"github.com/restgo/session"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
	"fmt"
)

type MongoSessionStore struct {
	maxAge       int64
	dialInfo     *mgo.DialInfo
	dbDialConfig *mongoConfig
	client       *mgo.Session
}

// mongodb session model
type dbSessionModel struct {
	Sid      string                 `json:"sid"`
	ExpireAt int64                  `json:"expireat"`
	Data     map[string]interface{} `json:"data"`
}

type mongoConfig struct {
	Hosts      string `json:"Hosts"`
	Database   string `json:"Database"`
	Collection string `json:"Collection,emitemgpty"`
	Username   string `json:"Username"`
	Password   string `json:"Password"`
}

// create  MongoSessionStore instance
func NewMongoSessionStore(options string) *MongoSessionStore {
	mongo := &MongoSessionStore{}
	err := json.Unmarshal([]byte(options), &mongo.dbDialConfig)
	if err != nil {
		panic("Read mongo session store options error: " + options)
	}
	if mongo.dbDialConfig.Collection == "" {
		mongo.dbDialConfig.Collection = "sessions"
	}

	mongo.dialInfo = &mgo.DialInfo{
		Addrs:    strings.Split(mongo.dbDialConfig.Hosts, ","),
		Timeout:  60 * time.Second,
		Database: mongo.dbDialConfig.Database,
		Username: mongo.dbDialConfig.Username,
		Password: mongo.dbDialConfig.Password,
	}
	mongo.client, err = mgo.DialWithInfo(mongo.dialInfo)
	if err == nil {
		// create expire collection
		err = mongo.client.DB(mongo.dbDialConfig.Database).C(mongo.dbDialConfig.Collection).EnsureIndex(mgo.Index{
			Key:         []string{"expireat"},
			ExpireAfter: time.Duration(0),
		})
	} else {
		panic("Connect to Mongo DB err: " + err.Error())
	}

	return mongo
}

func (this *MongoSessionStore) Init(sessionOptions string) error {
	jh := jsonhelper.NewJsonHelper([]byte(sessionOptions))
	this.maxAge = jh.Int64("MaxAge", 86400)

	return nil
}

func (this *MongoSessionStore) Get(sid interface{}) (*session.Session, error) {
	if this.client == nil {
		if err := this.connectInit(); err != nil {
			return nil, err
		}
	}
	realSid, ok := sid.(string)
	if !ok {
		return nil, fmt.Errorf("sid it not a string")
	}

	values, err := this.get(realSid)
	if err == mgo.ErrNotFound {
		// not exist, create a new sid for new session
		objId := bson.NewObjectId()
		ss := session.NewSession(this, objId.Hex(), make(map[string]interface{}))
		return ss, nil
	}
	ss := session.NewSession(this, realSid, values)
	return ss, nil
}

func (this *MongoSessionStore) Save(session *session.Session) (interface{}, error) {
	_, err := this.client.DB(this.dbDialConfig.Database).C(this.dbDialConfig.Collection).Upsert(bson.M{"sid": session.Sid}, dbSessionModel{
		session.Sid,
		time.Now().Unix() + this.maxAge, // seconds
		session.Values,
	})
	return session.Sid, err
}

func (this *MongoSessionStore) Destroy(sid interface{}) error {
	if this.client == nil {
		if err := this.connectInit(); err != nil {
			return err
		}
	}
	realSid, ok := sid.(string)
	if !ok {
		return fmt.Errorf("sid it not a string")
	}
	return this.del(realSid)
}

func (this *MongoSessionStore) StoreName() string {
	return "mongo"
}

// SessionDestroy delete mongodb session by id
func (this *MongoSessionStore) SessionDestroy(sid string) error {
	if this.client == nil {
		if err := this.connectInit(); err != nil {
			return err
		}
	}
	return this.del(sid)
}

// internal shortcut functions
func (this *MongoSessionStore) connectInit() error {
	var err error
	this.client, err = mgo.DialWithInfo(this.dialInfo)
	return err
}

func (this *MongoSessionStore) get(sid string) (map[string]interface{}, error) {
	var model dbSessionModel
	err := this.client.DB(this.dbDialConfig.Database).C(this.dbDialConfig.Collection).Find(bson.M{"sid": sid}).One(&model)
	if err != nil {
		return nil, err
	}

	return model.Data, nil
}

func (this *MongoSessionStore) set(sid string, data map[string]interface{}) error {
	_, err := this.client.DB(this.dbDialConfig.Database).C(this.dbDialConfig.Collection).Upsert(bson.M{"sid": sid}, dbSessionModel{
		sid,
		time.Now().Unix() + this.maxAge, //seconds
		data,
	})

	return err
}

func (this *MongoSessionStore) del(sid string) error {
	return this.client.DB(this.dbDialConfig.Database).C(this.dbDialConfig.Collection).Remove(bson.M{"sid": sid})
}
