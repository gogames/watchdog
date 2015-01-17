watchdog
----

Monitoring servers by sending ICMP Echo request from different servers to the monitored servers.

Take a tour here. http://mywatchdog.org/

### front-end

I am new to html/js/css stuff, so I simply copy and modify http://lab2023.github.io/hierapolis/

### main-server

- Written in [**Golang**](http://golang.org)
- Manage the ping-nodes running on different servers in varies location
- Store data
- Provide APIs to front-end
- Thanks to [**hprose**](https://github.com/hprose/hprose-go) and [**beego**](https://github.com/astaxie/beego)
- A lot more

### ping-node
- Written in [**Golang**](http://golang.org)
- Simply ping, same as ping command, rely on [**ping**](https://github.com/gogames/ping)

### TODO

- Documentation
- Makefile
- ....
