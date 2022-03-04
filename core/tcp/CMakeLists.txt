cmake_minimum_required(VERSION 2.8)
project(opentcp C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_VERBOSE_MAKEFILE ON)

include_directories(include)

file(GLOB SOURCES "src/*.c")

add_executable(opentcp main.c ${SOURCES})

target_link_libraries(opentcp pthread)
