package services

//type service struct{}
//
//func (service *service) createEvent(eventType string, source string, payload any) (cloudevents.Event, error) {
//	event := cloudevents.NewEvent()
//
//	event.SetSource(source)
//	event.SetType(eventType)
//	event.SetTime(time.Now().UTC())
//	event.SetID(uuid.New().String())
//
//	if err := event.SetData(cloudevents.ApplicationJSON, payload); err != nil {
//		msg := fmt.Sprintf("cannot encode %T [%#+v] as JSON", payload, payload)
//		return event, stacktrace.Propagate(err, msg)
//	}
//
//	return event, nil
//}
