# stationmaster
Software for running a ham radio station (for the database details see the end
of this file)

Functions:
1.  Morse code sending tutor
2.  Morse code keyer
3.  Station contact logger - log editor
4.  Analysis functions to review contact progress
5.  Building ADI files for upload to the LOTW and updating logs from downloads
6.  A contest mode where all the focus is on contesting
7.  Building Cabrillo files for submitting contest logs
8.  Interface to DX Spiders and display of active users on the VFO page
9.  Interface to the Ten Tec band switch to automatically know the band
10. Interface to a frequency synthesizer to control the radio frequency
11. Interface to the WSJT-X application for logging FT8/FT4 contacts
12. Interface to QRZ.com for pulling in contact information 

If you want to use the keyer and tutor functionality, you need a Raspberry Pi
and also to build the described hardware (or something like it).
Otherwise, you can run the program on just about any Mac or Windows machine.

My Ten Tec Omoni D radio does not have a CAT interface.  So, I read its band
switch (intended for a linear amp) so the software can follow the radio as
the bands are switched.  I also drive the external VFO input (the shorting
bar on the radio from the VFO out to VFO in is cut) with a digitally synthesized
VFO.  So, the software can control the radio frequency.  The software also
has hard limits at the two end of each ham band and knows the phone, CW,
FT8 and FT4 frequencies.

This program for the most part follows the patterns described in Alex Edward's
book "Let's Go".  If you are planning to deploy or modify this program, I
recommend getting a copy and reading it through.

There are plenty of resources on the web for copying Morse code but none for
the practice of sending Morse code.  I hope this fills the need, it does for me.  Using the keyer
and the tutor requires hardware and I have it running on a Raspberry Pi (RPi).  
The rest of the application can run anywhere (there are two different shell
files for RPi (rpi.sh) and not RPi (run.sh) usage.   

The Morse code subsystem uses the Farnsworth algorithm.  You can Google it and
find the old QST magazine article.    

I have looked around and have not found a contact manager that I liked for Mac.  
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
It is temperature compensated, but it has its limits.

A few notes on the logger:
1. The callsign search window let's you search for the call sign prior to a QSO,
get info about a contact, and find out if you have worked them before?
2. The add button (next to the logger button) on the top ribbon is how you add
a QSO.  You only see the add button in you are in the logger mode.
3. If you have set the defaults (using the default button on the top ribbon),
the band and mode show up on the add window.
4. After you move the curser out of the call sign window, the system looks up
the call sign from the local database or from QRZ.com and update the add window.
5. You only need to add the RST sent and received and any comments.
6. If you need to edit the entry, click on the ID link.
7. If you need to see more detail on the contact, click on the call sign.
8. When the application makes an API call to QRZ, it stores all of the information
obtained from QRZ in the database.  It only displays some of that information.
Some of it is used later in features like county summary.  The rest is just stored
for future analysis.
9. In the default page, in addition to the band and mode, you can set the contest
mode, the contest name, and the contest exchanges.  When you are in contest mode,
it will add them to the add page so you don't have to type them in.
10. When in the contest mode, non contest logs are hidden.

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

The Cabrillo button brings up the Cabrillo file generation page. All the fields
on this page are required.  All dates and times are in UTC.  The file is stored
in the contest directory as specified in the config.yaml file.

A few final notes:
1. I am building this as a single user local application
2. It is intended to run on a Raspberry Pi as the station controller
3. It also depends on specific hardware such as key paddle, oscillator and amplifier
4. It has long running programs (e.g. the keyer and key tutor)

Hence I have taken some liberties with how I have used the Go web
libraries.  If this was a multi-user application meant to be accessible
on the Internet, you would not do some of these and add other functionality
that I have not included.

Database installation (for MAC).

This is my way.  You may have better ways in which case you can use them.

1.  Make sure you have homebrew installed
2.  Run "brew install mysql" (it installs mysql without password which fine)
3.  To restart mysql at each login run "brew services start mysql"
4.  If you want to secure the installaton run "mysql_secure_installation" (I have not)
5.  Run "mysql -u root" now you are at the mysql prompt (note that root user has access)
6.  You can run "SHOW DATABASES;" and see that there are only native system databases there.
7.  Create the "stationmaster" database by running:

CREATE DATABASE stationmaster CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

8.  And run "SHOW DATABASES" again and you should see stationmaster in the list.
9.  Run "USE stationmaster;" and you will switch to the stationmaster database
10. Run "SHOW TABLES;" and you should see an empty set
11. Now you are ready to build the four tables: defaults, qrztable, stashtable, and stationlog 
12. See the table schemas below to see how these tables are structured.
13. Run "source makelogtable.txt;" to build the stationlogs table
14. Run "source makeqrztable.txt;" to build the qrztable
15. Run "source makestashtable.txt;" to build stashtable
16. Run "source makedefaulttable.txt;" to build defaults table
15. Create user by running "CREATE USER 'web'@'localhost';"
16. Give user permiissions by running: 

"GRANT SELECT, INSERT, UPDATE, DELETE ON stationmaster.* TO 'web'@'localhost';"

17. Set password for the user (use a password of your choosing instead of "password"

ALTER USER 'web'@'localhost' IDENTIFIED BY 'password';

18. Init the stash table by running "source initstash.txt;"
19. Quit mysql
20. Init defaults table by running (at the os prompt) "mysql -u root stationmaster < initdefaults.sql"
21. Set the environment variable STATIONMASTER to the root directory of the project.
22. For example in ~/.zshrc

STATIONMASTER=$HOME/Documents/gocode/src/stationmaster
export STATIONMASTER

Now you can run the program.



STATIONMASTER TABLE STRUCTURE

This is the schema for the stationlogs table.  The file makelogstable.txt in the dbscripts folder
generates this table when run in batch mode.

                                                                                  
| Field       | Type         | Null | Key | Default             | Extra          |
|-------------|--------------|------|-----|---------------------|----------------|
| id          | int(11)      | NO   | PRI | NULL                | auto_increment |
| time        | datetime     | NO   | MUL | NULL                |                |
| callsign    | varchar(20)  | NO   | MUL | NULL                |                |
| mode        | varchar(20)  | NO   |     | NULL                |                |
| sent        | varchar(10)  | NO   |     | NULL                |                |
| rcvd        | varchar(10)  | NO   |     | NULL                |                |
| band        | varchar(10)  | NO   |     | NULL                |                |
| name        | varchar(100) | NO   |     | NULL                |                |
| country     | varchar(100) | NO   |     | NULL                |                |
| comment     | varchar(100) | NO   |     | NULL                |                |
| lotwsent    | varchar(20)  | NO   |     | NULL                |                |
| lotwrcvd    | varchar(20)  | NO   |     | NULL                |                |
| lotwqsodate | datetime     | NO   |     | 1970-01-02 00:00:00 |                |
| lotwqsldate | datetime     | NO   |     | 1970-01-02 00:00:00 |                |
| contest     | varchar(5)   | NO   |     | No                  |                |
| exchsent    | varchar(10)  | NO   |     |                     |                |
| exchrcvd    | varchar(10)  | NO   |     |                     |                |
| contestname | varchar(50)  | NO   |     |                     |                |
                                                                                  

If you note, I store very little user information in the stationlogs table (I should
 store nothing but the call sign).  The detail information is stored in the qrztable
and here is the schema (note tht qso_count is not QRZ.com data, it is simply the
count of how many QSOs I have had with this contact.

Note in the script for building this table for MariaDB, there is no foreign key since
MariaDB did not support it when I first built the tables (and it may still not support it).
But the software does not depend on it.

                                                                       
| Field        | Type         | Null | Key | Default | Extra          |
|--------------|--------------|------|-----|---------|----------------|
| id           | int(11)      | NO   | PRI | NULL    | auto_increment |
| time         | datetime     | NO   |     | NULL    |                |
| callsign     | varchar(20)  | NO   | UNI | NULL    |                |
| aliases      | varchar(50)  | NO   |     | NULL    |                |
| dxcc         | varchar(5)   | NO   |     | NULL    |                |
| first_name   | varchar(100) | NO   |     | NULL    |                |
| last_name    | varchar(100) | NO   |     | NULL    |                |
| nickname     | varchar(50)  | NO   |     | NULL    |                |
| born         | varchar(5)   | NO   |     | NULL    |                |
| addr1        | varchar(50)  | NO   |     | NULL    |                |
| addr2        | varchar(50)  | NO   |     | NULL    |                |
| state        | varchar(20)  | NO   |     | NULL    |                |
| zip          | varchar(10)  | NO   |     | NULL    |                |
| country      | varchar(50)  | NO   |     | NULL    |                |
| country_code | varchar(5)   | NO   |     | NULL    |                |
| lat          | varchar(15)  | NO   |     | NULL    |                |
| lon          | varchar(15)  | NO   |     | NULL    |                |
| grid         | varchar(10)  | NO   |     | NULL    |                |
| county       | varchar(50)  | NO   |     | NULL    |                |
| fips         | varchar(10)  | NO   |     | NULL    |                |
| land         | varchar(50)  | NO   |     | NULL    |                |
| cqzone       | varchar(5)   | NO   |     | NULL    |                |
| ituzone      | varchar(5)   | NO   |     | NULL    |                |
| geolocation  | varchar(10)  | NO   |     | NULL    |                |
| effdate      | varchar(10)  | NO   |     | NULL    |                |
| expdate      | varchar(10)  | NO   |     | NULL    |                |
| prevcall     | varchar(10)  | NO   |     | NULL    |                |
| class        | varchar(5)   | NO   |     | NULL    |                |
| codes        | varchar(5)   | NO   |     | NULL    |                |
| qslmgr       | varchar(100) | NO   |     | NULL    |                |
| email        | varchar(50)  | NO   |     | NULL    |                |
| url          | varchar(50)  | NO   |     | NULL    |                |
| views        | varchar(20)  | NO   |     | NULL    |                |
| bio          | varchar(50)  | NO   |     | NULL    |                |
| image        | varchar(150) | NO   |     | NULL    |                |
| moddate      | varchar(30)  | NO   |     | NULL    |                |
| msa          | varchar(5)   | NO   |     | NULL    |                |
| areacode     | varchar(5)   | NO   |     | NULL    |                |
| timezone     | varchar(20)  | NO   |     | NULL    |                |
| gmtoffset    | varchar(5)   | NO   |     | NULL    |                |
| dst          | varchar(3)   | NO   |     | NULL    |                |
| eqsl         | varchar(3)   | NO   |     | NULL    |                |
| mqsl         | varchar(3)   | NO   |     | NULL    |                |
| attn         | varchar(100) | NO   |     | NULL    |                |
| qso_count    | int(11)      | NO   |     | NULL    |                |
                                                                       

The stashtable is identical to the qrztable and it is used to stash data
on a temporary basis to avoid multiple calls to qrz.com during one transaction.

The defaults table stores system data such as frequency by band, mode, fixed
constest exchanges,  and the like to make the user intrface a little nicer.  
It is a simple key value pair and the software uses it as needed by specifying
a key and associating a value with it.

                                                                
| Field | Type         | Null | Key | Default | Extra          |
|-------|--------------|------|-----|---------|----------------|
| id    | int(11)      | NO   | PRI | NULL    | auto_increment |
| kee   | varchar(20)  | NO   | MUL | NULL    |                |
| val   | varchar(100) | NO   |     | NULL    |                |
                                                                
