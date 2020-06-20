cmake_minimum_required(VERSION 2.8)
project(core C)

set(CMAKE_C_STANDARD 99)

add_executable(core main.c src/socket.c include/socket.h src/tuntap.c include/tuntap.h include/types.h src/epoll.c include/epoll.h)
target_link_libraries(core pthread)
