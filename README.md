Flash Gordon CLI
================

This is the command line tool for interacting with the Flash Gordon flash
burner that I built. It interacts with the board over serial and can upload,
dump, chip erase, and sector erase parallel flash chips such as:

 * SST39SF010A
 * SST39SF020A
 * SST39SF040
 * ... other chips with the same, common pinout

Hardware
--------

Flash Gordon is a shield that plugs into a common ATmega128a AVR breakout
board. You can often find boards like this on eBay or Aliexpress by searching
for "atmega128a". [Here's one example of the board in
question](https://www.ebay.com/itm/173100806719)

I am working on a second version of the board that will be free of the little
error mentioned in the tweet below on the first version. The software on the
AVR board is written in C and is to be uploaded with the Arduino environment
using [MegaCore](https://github.com/MCUdude/MegaCore).

<blockquote class="twitter-tweet"><p lang="en" dir="ltr">I&#39;m working on a
flash burner shield and software for the common ATmega128 breakout board. It
will support the common parallel 32-pin flash DIP pinout for e.g. SST39SF010A
or Am29F010B families (including the larger chips). Just using it as a learning
experience. <a
href="https://t.co/QUT4oCoo5p">pic.twitter.com/QUT4oCoo5p</a></p>&mdash; Karl
Matthias (@relistan) <a
href="https://twitter.com/relistan/status/1282256215960096775?ref_src=twsrc%5Etfw">July
12, 2020</a></blockquote>
![Flash Gordon1](./images/image1.png)

<blockquote class="twitter-tweet"><p lang="en" dir="ltr">It works! Software I
wrote for the breadboard version just worked. 😁Two little h/w mistakes but
probably not going to worry about it. You can see the bodge wire and cut trace.
Also not quite enough clearance for the power jack on the AVR board so I
removed it (don&#39;t need it). <a
href="https://t.co/LKhlerHqjA">pic.twitter.com/LKhlerHqjA</a></p>&mdash; Karl
Matthias (@relistan) <a
href="https://twitter.com/relistan/status/1287041625605185538?ref_src=twsrc%5Etfw">July
25, 2020</a></blockquote>
![Flash Gordon2](./images/image2.png)

Platforms Supported
-------------------

It has been tested on macOS and Linux. It should work on Windows and releases
will include Windows binaries, but I do not have a Windows machine to test
with.
