Go-utils Logger
===============

Log extends the standard log package to include support for the eight BSD style
log severity levels. It is drop in compatible with the standard log package and
requires no changes to your existing code.

This is a basic overview of using this package. For more detailed
specifications, please use `godoc` or the  [godoc.org]
(http://godoc.org/github.com/inhies/go-log) website.

Log Levels
----------

Below is list of the supported log levels. Note that EMERG is the highest level 
and debug is the lowest. If you set the level as NULL, no logging will be       
performed.                                                                      

* EMERG
* ALERT
* CRIT
* ERR
* WARNING
* NOTICE
* INFO
* DEBUG

Usage
-----

To utilize the log levels, add the following to your `import` statement:

    "github.com/inhies/go-log"
    _ "log"

Then initialize your logger with something like `l, err :=
log.NewLevel(log.ERR,true, os.Stdout, "", log.Ltime)` which will log all message
of ERR level and higher to Stdout and prepend the severity to the logged
message.

You can now log a message with `l.Err("This is an error message")` which will
be displayed on Stdout. If you were to try `l.Debug("This is a debug message")`
nothing would be displayed because DEBUG is a lower level than ERR. A CRIT,
ALERT, or EMERG message would still be displayed though.

You can change the log level at runtime by editing logger.Logger.Level. In
the example above, to switch from ERR to DEBUG level, you could do `l.Level =
log.DEBUG`.

Splitting Output
----------------

If you want to send your log messages to multiple places, you can register a
channel using the Split() function. You can choose to only have messages of your
specified level sent, or for all messages to be sent.

Formatting Messages
-------------------

Go-utils/log has support for matching the output format of log.Print,
log.Println, and log.Printf. To utilize these, you would use, for example,
log.Debug, log.Debugln, and log.Debugf, respectively.
