package schedule

import (
	"errors"
	"time"

	com "github.com/att/deadline/common"
)

// import (
// 	"bytes"
// 	"encoding/xml"
// 	"time"

// 	"github.com/att/deadline/dao"

// 	"github.com/att/deadline/common"
// 	"github.com/att/deadline/notifier"
// )

// func ConvertTime(timing string) time.Time {
// 	var m = int(time.Now().Month())
// 	loc, err := time.LoadLocation("Local")
// 	common.CheckError(err)
// 	parsedTime, err := time.ParseInLocation("15:04:05", timing, loc)
// 	if err != nil {
// 		parsedTime = time.Time{}
// 	}
// 	if !parsedTime.IsZero() {
// 		parsedTime = parsedTime.AddDate(time.Now().Year(), m-1, time.Now().Day()-1)
// 	}
// 	return parsedTime

// }

// func EvaluateSuccess(e *common.Event) bool {
// 	if !e.IsLive {
// 		return true
// 	}
// 	return e.Success
// }
// func EvaluateEvent(e *common.Event, h notifier.NotifyHandler) bool {
// 	return EvaluateTime(e, h) && EvaluateSuccess(e)
// }

// func (s *Schedule) EventOccurred(e *common.Event) {

// 	ev := findEvent(s.Start, e.Name)

// 	if ev != nil {
// 		ev.ReceiveAt = e.ReceiveAt
// 		ev.IsLive = true
// 		ev.Success = e.Success
// 		s.Start.OkTo = &s.End

// 	} else {
// 		s.Start.ErrorTo = &s.Error
// 	}

// }

// func MakeNodes(s *dao.ScheduleBlueprint) {
// 	fixSchedule(s)
// 	var f common.Event
// 	buf := bytes.NewBuffer(s.ScheduleContent)
// 	dec := xml.NewDecoder(buf)
// 	for dec.Decode(&f) == nil {
// 		e := f
// 		valid := e.ValidateEvent()
// 		if valid != nil {
// 			common.Debug.Println("You had an invalid event")
// 			return
// 		}
// 		node1 := common.Node{
// 			Event: &e,
// 			Nodes: []common.Node{},
// 		}
// 		s.Start.Nodes = append(s.Start.Nodes, node1)
// 	}
// }

// func fixSchedule(s *dao.ScheduleBlueprint) {
// 	evnts := []common.Event{}
// 	b := bytes.NewBuffer(s.ScheduleContent)
// 	d := xml.NewDecoder(b)

// 	for {
// 		t, err := d.Token()
// 		if err != nil {
// 			break
// 		}

// 		switch et := t.(type) {

// 		case xml.StartElement:
// 			if et.Name.Local == "event" {
// 				c := &common.Event{}
// 				if err := d.DecodeElement(&c, &et); err != nil {
// 					panic(err)
// 				}
// 				evnts = append(evnts, (*c))
// 			}
// 		case xml.EndElement:
// 			break
// 		}
// 	}
// 	bytes, err := xml.Marshal(evnts)
// 	common.CheckError(err)
// 	s.ScheduleContent = bytes
// }

func FromBlueprint(blueprint *com.ScheduleBlueprint) (*Schedule, error) {
	maps := &com.BlueprintMaps{}
	var err error = nil

	if maps, err = com.GetBlueprintMaps(blueprint); err != nil {
		return nil, err
	}

	schedule := &Schedule{
		nodes:         make(map[string]*NodeInstance),
		blueprintMaps: *maps,
	}

	schedule.End = &NodeInstance{
		NodeType: EndNodeType,
		value: &EndNode{
			name: blueprint.End.Name,
		},
	}

	if firstEvent, found := maps.Events[blueprint.Start.To]; !found {
		return nil, errors.New("Start node needs to point to an event Node")
	} else {
		schedule.addEventBlueprint(firstEvent)
	}

	// schedule.Start = &NodeInstance{
	// 	value: &StartNode{
	// 		to
	// 	}
	// }

	// handlerMap[handler.Name] = handler
	// if node := fromHandlerBlueprint(handler); node != nil {
	// 	if _, found := schedule.nodes[node.value.Name()]; found {
	// 		return nil, errors.New("Two or more nodes use the same name " + node.value.Name())
	// 	} else {
	// 		schedule.nodes[node.value.Name()] = node
	// 	}
	// }

	// for _, event := range blueprint.Events {
	// 	eventMap[event.Name] = event
	// 	if node, err := fromEventBlueprint(event); err != nil {
	// 		return nil, err
	// 	} else if _, found := schedule.nodes[node.value.Name()]; found {
	// 		return nil, errors.New("Two or more nodes use the same name " + node.value.Name())
	// 	} else {
	// 		schedule.nodes[node.value.Name()] = node
	// 	}
	// }

	// for _, node := range schedule.nodes {
	// 	nodeName := node.value.Name()
	// 	if nodeBlueprint, ok := eventMap[nodeName]; ok {

	// 	} else if nodeBlueprint, ok := handlerMap[nodeName]; ok {

	// 	}
	// }

	return schedule, nil
}

func (schedule *Schedule) addEventBlueprint(blueprint com.EventBlueprint) error {
	if c, err := com.FromBlueprint(time.Now(), blueprint.Constraints); err != nil {
		return err
	} else {

		// look for and make the okTo node
		if _, found := schedule.nodes[blueprint.OkTo]; !found {
			okToBlueprint, isEvent := schedule.blueprintMaps.Events[blueprint.OkTo]

			if isEvent { //okTo not already made and is an event node
				if err := schedule.addEventBlueprint(okToBlueprint); err != nil {
					return err
				}
			} else {
				return errors.New("Events can only ok-to other events")
			}
		}

		// look for and make the errorTo node
		if _, found := schedule.nodes[blueprint.ErrorTo]; !found {
			errorToBlueprint, isEvent := schedule.blueprintMaps.Events[blueprint.ErrorTo]

			if isEvent {
				if err := schedule.addEventBlueprint(errorToBlueprint); err != nil {
					return err
				}
			}

			errorToHandlerBlueprint, isHandler := schedule.blueprintMaps.Handlers[blueprint.ErrorTo]
			if isHandler {
				if err = schedule.addHandlerBlueprint(errorToHandlerBlueprint); err != nil {
					return err
				}
			} else {
				// at this point it wasn't found, and it wasn't an event and it wasn't a handler
				return errors.New("Couldn't find the error-to node for " + blueprint.Name)
			}

		}

		node := &NodeInstance{
			NodeType: EventNodeType,
			value: EventNode{
				name:        blueprint.Name,
				events:      make([]*com.Event, 0),
				constraints: c,
				okTo:        schedule.nodes[blueprint.OkTo],
				errorTo:     schedule.nodes[blueprint.ErrorTo],
			},
		}

		schedule.nodes[node.value.Name()] = node
		return nil
	}
}

func (schedule *Schedule) addHandlerBlueprint(blueprint com.HandlerBlueprint) error {

	if _, found := schedule.nodes[blueprint.To]; !found {
		okToEvent, isEvent := schedule.blueprintMaps.Events[blueprint.To]

		if isEvent {
			if err := schedule.addEventBlueprint(okToEvent); err != nil {
				return err
			}
		}

		okToHandler, isHandler := schedule.blueprintMaps.Handlers[blueprint.To]
		if isHandler {
			if err := schedule.addHandlerBlueprint(okToHandler); err != nil {
				return err
			}
		} else {
			// at this point it wasn't found, and it wasn't an event and it wasn't a handler
			return errors.New("Couldn't find the ok-to node for " + blueprint.Name)
		}
	}

	if blueprint.Email.EmailTo != "" {
		node := &NodeInstance{
			NodeType: HandlerNodeType,
			value: EmailHandlerNode{
				emailTo: blueprint.Email.EmailTo,
				to:      schedule.nodes[blueprint.To],
			},
		}

		schedule.nodes[node.value.Name()] = node
	} else {
		return errors.New("Handler " + blueprint.Name + " incorrectly defined.")
	}

	return nil
}
