package model

// CommentSchema 评论
type CommentSchema struct {
	ID          int64  `protobuf:"varint,1,opt,name=ID,proto3" json:"id"`                    // @gotags: json:"id" 评论ID
	ObjID       int64  `protobuf:"varint,2,opt,name=ObjID,proto3" json:"obj_id"`             // @gotags: json:"obj_id" 评论对象ID
	ObjType     int64  `protobuf:"varint,3,opt,name=ObjType,proto3" json:"obj_type"`         // @gotags: json:"obj_type" 评论对象类型
	MemberID    int64  `protobuf:"varint,4,opt,name=MemberID,proto3" json:"member_id"`       // @gotags: json:"member_id" 作者用户ID
	CommentID   int64  `protobuf:"varint,5,opt,name=CommentID,proto3" json:"comment_id"`     // @gotags: json:"comment_id" 同评论indx_id
	AtMemberIDs string `protobuf:"bytes,6,opt,name=AtMemberIDs,proto3" json:"at_member_ids"` // @gotags: json:"at_member_ids" at用户ID列表
	IP          string `protobuf:"bytes,7,opt,name=IP,proto3" json:"ip"`                     // @gotags: json:"ip" 评论IP
	Platform    int64  `protobuf:"varint,8,opt,name=Platform,proto3" json:"platform"`        // @gotags: json:"platform" 评论平台
	Device      string `protobuf:"bytes,9,opt,name=Device,proto3" json:"device"`             // @gotags: json:"device" 评论设备
	Message     string `protobuf:"bytes,10,opt,name=Message,proto3" json:"message"`          // @gotags: json:"message" 评论内容
	Meta        string `protobuf:"bytes,11,opt,name=Meta,proto3" json:"meta"`                // @gotags: json:"meta" 评论元数据 背景 字体
	ReplyID     int64  `protobuf:"varint,12,opt,name=ReplyID,proto3" json:"reply_id"`        // @gotags: json:"reply_id" 被回复的评论ID
	State       int64  `protobuf:"varint,13,opt,name=State,proto3" json:"state"`             // @gotags: json:"state" 评论状态 0-正常, 1-隐藏
	RootID      int64  `protobuf:"varint,14,opt,name=RootID,proto3" json:"root_id"`          // @gotags: json:"root_id" 根评论id 不为0则为回复一级评论
	CreatedAt   int64  `protobuf:"varint,15,opt,name=CreatedAt,proto3" json:"created_at"`    // @gotags: json:"created_at" 创建时间
	Floor       int64  `protobuf:"varint,16,opt,name=Floor,proto3" json:"floor"`             // @gotags: json:"floor" 楼层
	LikeCount   int64  `protobuf:"varint,17,opt,name=LikeCount,proto3" json:"like_count"`    // @gotags: json:"like_count" 点赞数
	HateCount   int64  `protobuf:"varint,18,opt,name=HateCount,proto3" json:"hate_count"`    // @gotags: json:"hate_count" 踩数
	Count       int64  `protobuf:"varint,19,opt,name=Count,proto3" json:"count"`             // @gotags: json:"count" 评论数
	RootCount   int64  `protobuf:"varint,20,opt,name=RootCount,proto3" json:"root_count"`    // @gotags: json:"root_count" 根评论数
}
