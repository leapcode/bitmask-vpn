// go:build cgo
// +go:build cgo

package backend

import (
	"sync"
	"unsafe"

	"github.com/rs/zerolog/log"
)

/* ATCHUNG! what follow are not silly comments. Well, *this one* is, but
   the lines after this are not.
   Those are inline C functions, that are invoked by CGO later on.
   it's also crucial that you don't any extra space between the function
   block and the 'import "C"' line. */

// typedef void (*cb)();
// inline void _do_callback(cb f) {
// 	f();
// }
import "C"

/* callbacks into C-land. We keep a registry, and protect its updates with a mutex. */

var callbacks = make(map[string](*[0]byte))
var callbackMutex sync.Mutex

// Events are just a enumeration of all the posible events that C functions can
// be interested in subscribing to. You cannot subscribe to an event that is
// not listed here.
type Events struct {
	OnStatusChanged string
}

const OnStatusChanged string = "OnStatusChanged"

// subscribe registers a callback from C-land.
// This callback needs to be passed as a void* C function pointer.
func subscribe(event string, fp unsafe.Pointer) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	/* I'm commenting this check because it imposes 1.14, which is only in buster-backports.
	   We can re-add it after buster is oldstable
	   e := &Events{}
	   v := reflect.Indirect(reflect.ValueOf(&e))
	   hf := v.Elem().FieldByName(event)
	   if reflect.ValueOf(hf).IsZero() {
	           fmt.Println("ERROR: not a valid event:", event)
	   } else {
	           callbacks[event] = (*[0]byte)(fp)
	   }
	*/
	callbacks[event] = (*[0]byte)(fp)
}

// trigger fires a callback from C-land.
func trigger(event string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	cb := callbacks[event]
	if cb != nil {
		C._do_callback(cb)
	} else {
		log.Warn().
			Str("event", event).
			Msg("This event does not have subscribers")
	}
}
