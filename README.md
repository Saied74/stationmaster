# stationmaster
Software for running a ham radio station

Functions:
1. Morse code sending tutor
2. Morse code keyer (not yet built)
3. Station contact logger - log editor
4. Antenna tuner (coming up after keyer)

There are plenty of resources on the web for copying code but none for
sending code.  I hope this fills the need, it does for me.  I run a
Ten Tec Omini D which does not have a keyer built in (more modern rigs
do).  This will fill this gap.  I have looked around and have not found
a contact manager that was simple and easy to use on a Mac.  So, I have
built a simple one.  I plan to enhance it with a few more functions but
never to complicate it.

This is what you need to use this application:
1. Raspberry Pi with enough memory to run Go and mySQL
2. I run with RPi native OS and it seems to work fine.
3. Go programming package
4. Gobot Raspi package
5. mySQL
6. Keyer - tutor hardware if you would like to use that functionality

Schematic for the keyer-tutor is in the resources directory.  It is
currently running on a breadboard.  Once I transfer it to a circuit board,
I will include a picture.

The application usage is self explanatory and breif notes are provide
in the application screens.

A few final notes:
1. I am building this as a single user local application
2. It is intended to run on a Raspberry Pi as the station controller
3. It also depends on specific hardware such as key paddle, oscillator and amplifier
4. It has long running programs (e.g. the keyer and key tutor)

Hence I have taken some liberties with how I have used the Go web
libraries.  If this was a multi-user application meant to be accessible
on the Internet, you would not do some of these and add other functionality
that I have not included.
