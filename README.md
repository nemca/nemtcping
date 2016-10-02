# TCPing
Yet another tcping tool.

Similar to https://github.com/hwsdien/gotcping and https://github.com/pjperez/gotcping

## Installing
    go get github.com/nemca/nemtcping

## Usage
    nemtcping [-c count] [-t timeout] [-p] [-q] <host> [<port>]

### Defaults
    -c     default to 0 requests, forever ping loop
    -t     default to 1 seconds
    <port> default to 80

### Example

#### Normal mode
    $ nemtcping github.com 443
    github.com port 443 open.
    $ nemtcping -t 2 github.com 444
    github.com port 444 closed.

#### Ping mode
    $ nemtcping -c 2 -p github.com 443
    nemtcping github.com (192.30.253.113)
    Connected to github.com:443, RTT=145.301 ms
    Connected to github.com:443, RTT=135.903 ms

    --- github.com nemtcping statistic ---
    2 packets transmitted, 2 packets received, 0.0% packet loss
    round-trip min/avg/max = 135.903/140.602/145.301 ms
