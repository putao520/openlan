cmake_minimum_required(VERSION 2.8)
project(core C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_VERBOSE_MAKEFILE ON)

include_directories(include)

add_executable(core main.c src/socket.c include/socket.h src/tuntap.c include/tuntap.h include/types.h src/control.c include/control.h src/message.c src/message.h)

target_link_libraries(core pthread)
