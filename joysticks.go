package joysticks

import (
	"math"
	"time"
	//	"fmt"
)

// TODO divided position events
// TODO drag event

var LongPressDelay = time.Second

type hatAxis struct {
	number   uint8
	axis     uint8
	reversed bool
	time     time.Duration
	value    float32
}

type button struct {
	number uint8
	time   time.Duration
	value  bool
}

//HID holds the in-coming event channel, mappings, and registered events for a device, and has methods to control and adjust behaviour.
type HID struct {
	OSEvents              chan osEventRecord
	Buttons               map[uint8]button
	HatAxes               map[uint8]hatAxis
	buttonChangeEvents    map[uint8]chan Event
	buttonCloseEvents     map[uint8]chan Event
	buttonOpenEvents      map[uint8]chan Event
	buttonLongPressEvents map[uint8]chan Event
	hatChangeEvents       map[uint8]chan Event
	hatPanXEvents         map[uint8]chan Event
	hatPanYEvents         map[uint8]chan Event
	hatPositionEvents     map[uint8]chan Event
	hatAngleEvents        map[uint8]chan Event
	hatRadiusEvents       map[uint8]chan Event
	hatCenteredEvents     map[uint8]chan Event
	hatEdgeEvents         map[uint8]chan Event
}

type Event interface {
	Moment() time.Duration
}

type when struct {
	time time.Duration
}

func (b when) Moment() time.Duration {
	return b.time
}

// button changed
type ButtonEvent struct {
	when
	number uint8
	value  bool
}

// hat changed
type HatEvent struct {
	when
	number uint8
	axis   uint8
	value  float32
}

// Hat Axis changed event, X,Y {-1...1}
type HatPositionEvent struct {
	when
	X, Y float32
}

// Hat Axis changed event, V {-1...1}
type HatPanXEvent struct {
	when
	V float32
}

// Hat Axis changed event, V {-1...1}
type HatPanYEvent struct {
	when
	V float32
}

// Hat angle changed event, Angle {-Pi...Pi}
type HatAngleEvent struct {
	when
	Angle float32
}

// Hat radius changed event, R {-1...1}
type HatRadiusEvent struct {
	when
	R float32
}

// ParcelOutEvents waits on the HID.OSEvent channel (so is blocking), then puts the required event(s), on any registered channel(s).
func (d HID) ParcelOutEvents() {
	for {
		if evt, ok := <-d.OSEvents; ok {
			switch evt.Type {
			case 1:
				b := d.Buttons[evt.Index]
				if c, ok := d.buttonChangeEvents[b.number]; ok {
					c <- ButtonEvent{when{toDuration(evt.Time)}, b.number, evt.Value == 1}
				}
				if evt.Value == 0 {
					if c, ok := d.buttonOpenEvents[b.number]; ok {
						c <- when{toDuration(evt.Time)}
					}
					if c, ok := d.buttonLongPressEvents[b.number]; ok {
						if toDuration(evt.Time) > b.time+LongPressDelay {
							c <- when{toDuration(evt.Time)}
						}
					}
				}
				if evt.Value == 1 {
					if c, ok := d.buttonCloseEvents[b.number]; ok {
						c <- when{toDuration(evt.Time)}
					}
				}
				d.Buttons[evt.Index] = button{b.number, toDuration(evt.Time), evt.Value != 0}
			case 2:
				h := d.HatAxes[evt.Index]
				v := float32(evt.Value) / maxValue
				if h.reversed {
					v = -v
				}
				if c, ok := d.hatChangeEvents[h.number]; ok {
					c <- HatEvent{when{toDuration(evt.Time)}, h.number, h.axis, v}
				}
				switch h.axis {
				case 1:
					if c, ok := d.hatPanXEvents[h.number]; ok {
						c <- HatPanXEvent{when{toDuration(evt.Time)}, v}
					}
				case 2:
					if c, ok := d.hatPanYEvents[h.number]; ok {
						c <- HatPanYEvent{when{toDuration(evt.Time)}, v}
					}
				}
				if c, ok := d.hatPositionEvents[h.number]; ok {
					switch h.axis {
					case 1:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, v, d.HatAxes[evt.Index+1].value}
					case 2:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, d.HatAxes[evt.Index-1].value, v}
					}
				}
				if c, ok := d.hatAngleEvents[h.number]; ok {
					switch h.axis {
					case 1:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index+1].value)))}
					case 2:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index-1].value), float64(v)))}
					}
				}
				if c, ok := d.hatRadiusEvents[h.number]; ok {
					switch h.axis {
					case 1:
						c <- HatRadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(v)*float64(v) + float64(d.HatAxes[evt.Index+1].value)*float64(d.HatAxes[evt.Index+1].value)))}
					case 2:
						c <- HatRadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(d.HatAxes[evt.Index-1].value)*float64(d.HatAxes[evt.Index-1].value) + float64(v)*float64(v)))}
					}
				}
				if c, ok := d.hatEdgeEvents[h.number]; ok {
					// fmt.Println(v,h)
					if (v == 1 || v == -1) && h.value != 1 && h.value != -1 {
						switch h.axis {
						case 1:
							c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index+1].value)))}
						case 2:
							c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index-1].value), float64(v)))}
						}
					}
				}
				if c, ok := d.hatCenteredEvents[h.number]; ok {
					if v == 0 && h.value != 0 {
						switch h.axis {
						case 1:
							if d.HatAxes[evt.Index+1].value==0 {
								c <- when{toDuration(evt.Time)}
							}
						case 2:
							if d.HatAxes[evt.Index-1].value==0 {
								c <- when{toDuration(evt.Time)}
							}
						}
					}
				}
				d.HatAxes[evt.Index] = hatAxis{h.number, h.axis, h.reversed, toDuration(evt.Time), v}
			default:
				// log.Println("unknown input type. ",evt.Type & 0x7f)
			}
		} else {
			break
		}
	}
}

// Type of registerable methods and the index they are called with. (Note: the event type is indicated by the method.)
type Channel struct {
	Number uint8
	Method func(HID, uint8) chan Event
}

// button chnages
func (d HID) OnButton(button uint8) chan Event {
	c := make(chan Event)
	d.buttonChangeEvents[button] = c
	return c
}

// button goes open
func (d HID) OnOpen(button uint8) chan Event {
	c := make(chan Event)
	d.buttonOpenEvents[button] = c
	return c
}

// button goes closed
func (d HID) OnClose(button uint8) chan Event {
	c := make(chan Event)
	d.buttonCloseEvents[button] = c
	return c
}

// button goes open and the previous event, closed, was more than LongPressDelay ago.
func (d HID) OnLong(button uint8) chan Event {
	c := make(chan Event)
	d.buttonLongPressEvents[button] = c
	return c
}

// hat moved
func (d HID) OnHat(hat uint8) chan Event {
	c := make(chan Event)
	d.hatChangeEvents[hat] = c
	return c
}

// hat position changed
func (d HID) OnMove(hat uint8) chan Event {
	c := make(chan Event)
	d.hatPositionEvents[hat] = c
	return c
}

// hat axis-X moved
func (d HID) OnPanX(hat uint8) chan Event {
	c := make(chan Event)
	d.hatPanXEvents[hat] = c
	return c
}

// hat axis-Y moved
func (d HID) OnPanY(hat uint8) chan Event {
	c := make(chan Event)
	d.hatPanYEvents[hat] = c
	return c
}

// hat angle changed
func (d HID) OnRotate(hat uint8) chan Event {
	c := make(chan Event)
	d.hatAngleEvents[hat] = c
	return c
}

// hat moved to center
func (d HID) OnCenter(hat uint8) chan Event {
	c := make(chan Event)
	d.hatCenteredEvents[hat] = c
	return c
}

// hat moved to edge
func (d HID) OnEdge(hat uint8) chan Event {
	c := make(chan Event)
	d.hatEdgeEvents[hat] = c
	return c
}

// see if Button exists.
func (d HID) ButtonExists(button uint8) (ok bool) {
	for _, v := range d.Buttons {
		if v.number == button {
			return true
		}
	}
	return
}

// see if Hat exists.
func (d HID) HatExists(hat uint8) (ok bool) {
	for _, v := range d.HatAxes {
		if v.number == hat {
			return true
		}
	}
	return
}

// Button current state.
func (d HID) ReadButtonState(button uint8) bool {
	return d.Buttons[button].value
}

// Hat latest position. (coords slice needs to be long enough to hold all axis.)
func (d HID) ReadHatPosition(hat uint8, coords []float32) {
	for _, h := range d.HatAxes {
		if h.number == hat {
			coords[h.axis-1] = h.value
		}
	}
	return
}

// insert events as if from hardware.
func (d HID) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	d.OSEvents <- osEventRecord{Value: v, Type: t, Index: i}
}
