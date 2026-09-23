package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/util"
)

// Seed 初始化演示数据：管理员/演示用户/宠物/话题/动态/约伴。
func Seed(ctx context.Context, db *mongo.Database, logger *slog.Logger) error {
	logger.Info(constants.LogSeedStarted)
	users := db.Collection("users")
	adminCount, _ := users.CountDocuments(ctx, bson.M{"username": "admin"})
	if adminCount == 0 {
		adminHash, _ := util.HashPassword("admin123")
		admin := &model.User{
			Username:     "admin",
			PasswordHash: adminHash,
			Nickname:     "社区管理员",
			City:         "上海",
			Role:         constants.RoleAdmin,
			Status:       constants.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if _, err := users.InsertOne(ctx, admin); err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}
	}
	demoCount, _ := users.CountDocuments(ctx, bson.M{"username": "demo"})
	if demoCount == 0 {
		demoHash, _ := util.HashPassword("demo123")
		demo := &model.User{
			Username:     "demo",
			PasswordHash: demoHash,
			Nickname:     "铲屎官小萌",
			Avatar:       "https://images.unsplash.com/photo-1543466835-00a7907e9de1?w=200",
			Bio:          "家有柴犬一只，喜欢遛狗和拍狗子",
			City:         "上海",
			Role:         constants.RoleUser,
			Status:       constants.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if _, err := users.InsertOne(ctx, demo); err != nil {
			return fmt.Errorf("seed demo: %w", err)
		}
	}
	// 宠物
	petColl := db.Collection("pets")
	if n, _ := petColl.CountDocuments(ctx, bson.M{}); n == 0 {
		var demoUser model.User
		_ = users.FindOne(ctx, bson.M{"username": "demo"}).Decode(&demoUser)
		base := time.Date(2021, 6, 1, 0, 0, 0, 0, time.Local)
		pets := []model.Pet{
			{OwnerID: demoUser.ID, Name: "柴柴", Species: constants.PetSpeciesDog, Breed: "柴犬", Birthday: base, Gender: constants.PetGenderMale, Personality: "粘人、贪吃、爱撒娇", Bio: "一只快乐的小柴犬", Avatar: "https://images.unsplash.com/photo-1561037404-61cd46aa615b?w=400", City: "上海", Status: constants.PetStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{OwnerID: demoUser.ID, Name: "布丁", Species: constants.PetSpeciesCat, Breed: "英短", Birthday: base.AddDate(-2, 0, 0), Gender: constants.PetGenderFemale, Personality: "高冷但粘人", Bio: "傲娇小猫咪", Avatar: "https://images.unsplash.com/photo-1573865526739-10659fec78a5?w=400", City: "上海", Status: constants.PetStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{OwnerID: demoUser.ID, Name: "旺财", Species: constants.PetSpeciesDog, Breed: "金毛", Birthday: base.AddDate(-3, 0, 0), Gender: constants.PetGenderMale, Personality: "温柔大暖男", Bio: "金毛大暖男", Avatar: "https://images.unsplash.com/photo-1552053831-71594a27632d?w=400", City: "北京", Status: constants.PetStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		for i := range pets {
			if _, err := petColl.InsertOne(ctx, pets[i]); err != nil {
				return fmt.Errorf("seed pet: %w", err)
			}
		}
	}
	// 话题
	topicColl := db.Collection("topics")
	if n, _ := topicColl.CountDocuments(ctx, bson.M{}); n == 0 {
		topics := []model.Topic{
			{Name: "柴犬日常", Description: "分享柴犬的可爱日常", Status: constants.TopicStatusActive, CreatedAt: time.Now()},
			{Name: "猫咪治愈", Description: "猫咪治愈系瞬间", Status: constants.TopicStatusActive, CreatedAt: time.Now()},
			{Name: "遛狗打卡", Description: "每天遛狗打卡记录", Status: constants.TopicStatusActive, CreatedAt: time.Now()},
			{Name: "宠物成长记", Description: "记录宠物成长的点滴", Status: constants.TopicStatusActive, CreatedAt: time.Now()},
		}
		for i := range topics {
			if _, err := topicColl.InsertOne(ctx, topics[i]); err != nil {
				return fmt.Errorf("seed topic: %w", err)
			}
		}
	}
	// 动态
	postColl := db.Collection("posts")
	if n, _ := postColl.CountDocuments(ctx, bson.M{}); n == 0 {
		var demoUser model.User
		_ = users.FindOne(ctx, bson.M{"username": "demo"}).Decode(&demoUser)
		var pets []model.Pet
		cursor, _ := petColl.Find(ctx, bson.M{})
		_ = cursor.All(ctx, &pets)
		var petIDs []primitive.ObjectID
		if len(pets) > 0 {
			petIDs = []primitive.ObjectID{pets[0].ID}
		}
		posts := []model.Post{
			{AuthorID: demoUser.ID, PetIDs: petIDs, Content: "今天带柴柴去公园散步，超开心！#柴犬日常 #遛狗打卡", Type: constants.PostTypeImage, Media: []model.MediaItem{{Type: "image", URL: "https://images.unsplash.com/photo-1561037404-61cd46aa615b?w=800"}}, Topics: []string{"柴犬日常", "遛狗打卡"}, City: "上海", Location: "世纪公园", Status: constants.PostStatusApproved, LikeCount: 12, CommentCount: 3, FavoriteCount: 5, ForwardCount: 2, CreatedAt: time.Now().Add(-2 * time.Hour), UpdatedAt: time.Now()},
			{AuthorID: demoUser.ID, PetIDs: petIDs, Content: "布丁今天好粘人，一直求抱抱 #猫咪治愈", Type: constants.PostTypeImage, Media: []model.MediaItem{{Type: "image", URL: "https://images.unsplash.com/photo-1573865526739-10659fec78a5?w=800"}}, Topics: []string{"猫咪治愈"}, City: "上海", Location: "家", Status: constants.PostStatusApproved, LikeCount: 8, CommentCount: 2, FavoriteCount: 3, CreatedAt: time.Now().Add(-5 * time.Hour), UpdatedAt: time.Now()},
			{AuthorID: demoUser.ID, PetIDs: petIDs, Content: "金毛旺财的成长记录，从小奶狗到暖男 #宠物成长记", Type: constants.PostTypeImage, Media: []model.MediaItem{{Type: "image", URL: "https://images.unsplash.com/photo-1552053831-71594a27632d?w=800"}}, Topics: []string{"宠物成长记"}, City: "北京", Location: "朝阳公园", Status: constants.PostStatusApproved, LikeCount: 20, CommentCount: 5, FavoriteCount: 9, ForwardCount: 4, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		}
		for i := range posts {
			if _, err := postColl.InsertOne(ctx, posts[i]); err != nil {
				return fmt.Errorf("seed post: %w", err)
			}
			for _, t := range posts[i].Topics {
				_, _ = topicColl.UpdateOne(ctx, bson.M{"name": t}, bson.M{"$inc": bson.M{"post_count": 1}})
			}
		}
	}
	// 约伴
	meetupColl := db.Collection("meetups")
	if n, _ := meetupColl.CountDocuments(ctx, bson.M{}); n == 0 {
		var demoUser model.User
		_ = users.FindOne(ctx, bson.M{"username": "demo"}).Decode(&demoUser)
		meetups := []model.Meetup{
			{CreatorID: demoUser.ID, Title: "周六世纪公园遛狗局", Description: "带狗狗一起撒欢，欢迎柴犬柯基金毛", City: "上海", Location: "世纪公园2号门", MeetTime: time.Now().Add(72 * time.Hour), DurationMinutes: 120, MaxPeople: 8, Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{CreatorID: demoUser.ID, Title: "滨江夜跑遛狗", Description: "晚上一起遛狗跑步，附赠零食", City: "上海", Location: "徐汇滨江", MeetTime: time.Now().Add(120 * time.Hour), DurationMinutes: 90, MaxPeople: 5, Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		for i := range meetups {
			if _, err := meetupColl.InsertOne(ctx, meetups[i]); err != nil {
				return fmt.Errorf("seed meetup: %w", err)
			}
		}
	}
	// 私信示例
	chatColl := db.Collection("chat_messages")
	if n, _ := chatColl.CountDocuments(ctx, bson.M{}); n == 0 {
		var adminUser, demoUser model.User
		_ = users.FindOne(ctx, bson.M{"username": "admin"}).Decode(&adminUser)
		_ = users.FindOne(ctx, bson.M{"username": "demo"}).Decode(&demoUser)
		msgs := []model.ChatMessage{
			{FromID: demoUser.ID, ToID: adminUser.ID, Type: constants.ChatMessageText, Content: "你好呀，我是新来的铲屎官", Read: true, CreatedAt: time.Now().Add(-1 * time.Hour)},
			{FromID: adminUser.ID, ToID: demoUser.ID, Type: constants.ChatMessageText, Content: "欢迎加入宠物社区，有问题随时找我~", Read: false, CreatedAt: time.Now().Add(-59 * time.Minute)},
		}
		for i := range msgs {
			if _, err := chatColl.InsertOne(ctx, msgs[i]); err != nil {
				return fmt.Errorf("seed chat: %w", err)
			}
		}
	}
	logger.Info(constants.LogSeedDone)
	return nil
}
