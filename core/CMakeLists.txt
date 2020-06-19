cmake_minimum_required(VERSION 2.8)
project(core C)

set(CMAKE_C_STANDARD 99)

add_executable(core main.c tcp_server.c tcp_server.h)