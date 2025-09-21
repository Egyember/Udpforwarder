# Udpforwarder
A tool to forward udp packets to multiple hosts as unicast

# Building form source
## Install dependencies
```bash
go mod tidy
```
## Building
There are multiple ways to build it.
### basic executable
Run this command in the root directory:
```bash
go build
```
An executable binary will be created. For configuration it will use a file in the same directory. An example can be found in config.toml
### docker image
This requires make and docker buildx.  
1. set the architecture in the Makefile
2. Make sure the docker daemon is running
3. Run the following commands
   ```bash
   make build
   ```
if everything works as it should it will echo out all commands that it run and you will have a docker image with the namespecified in the Makefile with a tar subfix.
This option uses the config.toml in this repository for configuration. In the image it will be in the root directory if you want to overlay it.
The final image is around 4 mb in size and only contain the absolutely necessary files. (ie: no libc or gnu utils)

# configuration
Configuration uses a toml file in the same directory as the main executable.
## Global options
| name of option   | type   | Example   | what it dose   |
| --- | --- | --- | --- |
| syslog | boolean | true | whenever to send logs to a syslog server or not |
| logaddr | string | "192.168.3.110:514" | ip addres and port of the syslog server |
| rule | array of tables | see rules below | rules for forwarding packets |
| listen | table |  {ip = "", port = 6667}  | ip and port to listen on form packets |

## rules
| name of option   | type   | Example   | what it dose   |
| --- | --- | --- | --- |
| ip | string | "192.168.0.2" | target ip address |
| port | integer | 6667 | target port to forward packets |
| use-src | boolean | true | use the source address of the packet or send it from local address. This option requires net.ipv4.ip_nonlocal_bind = 1 kernel parameter on linux. Not tested on other platforms |

An example can be found in config.toml
