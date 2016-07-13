/*
Package joysticks, provides simplified event routing, through channels, from the Linux joystick driver File-like interface.

events can be listened for from any thread, re-routed and simulated.

usage:

'Capture', a single call to setup and start basic event routing.

or (more flexible)

'Connect' to a joystick by index number, then use methods to add event channels, one for each button or hat, and start running by calling 'ParcelOutEvents'.

event, an interface with a 'Moment' method, provides a time.Duration. Moment() returns whatever the underlying Linux driver provides as the events timestamp, in time.Duration.

hat channel event provides current position, (x,y) the event will need casting to the hat event to access these. (and only one axis actually changes per event.)

or (DIY)

'Connect' to a joystick by index number

handle all events directly using the returned Joystick's OSEvent channel.

*/
package joysticks

/*

could be used with little change on any linux 'input'

relies on initial event state burst to setup available events register.

linux driver doesn't allow blocking other programs from read from the same device, but they dont consume.

*/
