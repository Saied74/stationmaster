# stationmaster
Software for running a ham radio station

Functions:
1. Morse code sending tutor
2. Morse code keyer
3. Station contact logger - log editor
4. Analysis functions to review contact progress
5. Building ADI files for upload to the LOTW and updating logs from downloads

This program for the most part follows the patterns described in Alex Edward's
book "Let's Go".  If you are planning to deploy or modify this program, I
recommend getting a copy and reading it through.

There are plenty of resources on the web for copying code but none for
sending code.  I hope this fills the need, it does for me.  Using the keyer
and the tutor requires hardware and I have it running on a Raspberry Pi.  
The rest of the application can run anywhere (there are two different shell
files for RPi and not RPi usage.   

The Morse code subsystem uses the Farnsworth algorithm.  You can Google it and
find the old QSL magazine article.    

I have looked around and have not found a contact manager that I liked for a Mac.  
So, I have built this.  I plan to enhance it with a few more functions but
never to complicate it.  It also runs on Raspberry Pi or anywhere else that are
supported by Go, mySQL and standard browser functionality (e.g. html,
JavaScript and Bootstrap).

This is what you need to use this application:
1. Raspberry Pi with enough memory to run Go and mySQL, a Mac or Windows machine
2. I run with RPi native OS and Mac and it work fine.
3. Go programming package
4. Gobot Raspi package
5. mySQL
6. Keyer - tutor hardware if you would like to use that functionality
7. QRZ account

Schematic for the keyer-tutor is in the resources directory.  You can package
it as you like, but beware that the output is a class A amplifier and gets hot.

A few notes on the logger:
1. The callsign search window let's you search for the call sign prior to a QSO,
get info about a contact and find out if you have worked them before?
2. The add button (next to the logger button) on the top ribbon is how you add
a QSO.  You only see the add button in you are in the logger mode.
3. If you have set the defaults (using the default button on the top ribbon),
the band and mode show up on the add window.
4. After you move the curser out of the call sign window, the system looks up
the call sign from the local database or from QRZ and update the add window.
5. You only need to add the RST sent and received and any comments.
6. If you need to edit the entry, click on the ID link.
7. If you need to see more detail on the contact, click on the call sign.
8. When the application makes an API call to QRZ, it stores all of the information
obtained from QRZ in the database.  It only displays some of that information.
Some of it is used later in features like county summary.  The rest is just stored
for future analysis.

The analysis tab is all self explanatory.  I will be adding additional analytics
as the needs arise.

The ADIF button brings up the page for generating or updating LOTW status.
1. On the ADIF page, only logbook entries with blank LOTW Sent field are
displayed.  Once LOTW file is generated, this field is set to YES.
2. For uploading the ARRL ADIF file, put the filename window and push the
update QSL button.

The generated and uploaded ADIF files are in the the ADIF directory in the
configuration file.  The configuration chain works as follows:
1. The program depends on the directory structure as it is laid out.
2. The top of the configuration chain is the environment variable STATIONMASTER
3. STATIONMASTER points to the file config.yaml.  
4. The QSL log is referenced to the $HOME environment variable.   

A few final notes:
1. I am building this as a single user local application
2. It is intended to run on a Raspberry Pi as the station controller
3. It also depends on specific hardware such as key paddle, oscillator and amplifier
4. It has long running programs (e.g. the keyer and key tutor)

Hence I have taken some liberties with how I have used the Go web
libraries.  If this was a multi-user application meant to be accessible
on the Internet, you would not do some of these and add other functionality
that I have not included.
