package runner

func emit(task Task, eventType string, message string, payloadJSON string) {
	if task.OnEvent != nil {
		task.OnEvent(Event{EventType: eventType, Message: message, PayloadJSON: payloadJSON})
	}
}
