cmake_minimum_required(VERSION 2.8)
project(openudp C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_VERBOSE_MAKEFILE ON)

include_directories(include)
include_directories(../ovs/lib)

add_executable(openudp ./src/udp.c)

target_link_libraries(openudp pthread openvswitch)
