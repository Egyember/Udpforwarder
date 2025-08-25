# Udpforwarder
A tool to forward udp packets to multiple hosts as unicast

# Building
##Install dependencies
```bash
go mod tidy
```
## Building
There are multiple ways to build it.
### basic executible
Run this command in the root directory:
```bash
go build
```
An executible binary will be created. For configuration it will use a file in the same directory. An example can be found in config.toml
### docker image
This requires make.  
1. set the architucture in the Makefile
2. Make sure the docker daemon is running
3. Run the following commands
   ```bash
   make build
   ```
if everything works as it should it will echo out all commands that it run and you will have a docker image with the name speciflyed in the Makefile with a tar subfix.
This option uses the config.toml in this repository for configuration. In the image it will be in the root directory if you whant to overlay it.
The filale image is around 4 mb in size and only contain the absolutely nesesery files. (ie: no libc or gnu utils)

# configuration
Configuration uses a toml file in the same directory as the main executable.
## Global options
| name of option   | type   | Example   | what it dose   |
| --- | --- | --- | --- |
| log | boolian | true | whenever to do logs to a syslog server or not |
| logaddr | string | "192.168.3.110:514" | ip addres and port of the syslog server |
| rule | array of tables | see rules below | rules for forwarding packets |

## rules
todo
