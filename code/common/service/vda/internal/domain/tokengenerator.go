package domain

// TokenGenerator 抽象 LiveKit token 生成，便于测试时注入 stub。
// *livekit.Client 已实现此接口。
type TokenGenerator interface {
	GenerateToken(roomName string, userID int64, userName string) (string, error)
}
