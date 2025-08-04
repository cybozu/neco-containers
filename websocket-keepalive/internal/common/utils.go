package common

func WebSocketMessageType(n int) string {
	switch n {
	case 1:
		return "Text"
	case 2:
		return "Binary"
	case 8:
		return "Close"
	case 9:
		return "Ping"
	case 10:
		return "Pong"
	default:
		return "Unknown"
	}
}
