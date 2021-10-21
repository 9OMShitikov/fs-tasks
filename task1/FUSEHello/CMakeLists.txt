cmake_minimum_required(VERSION 3.16)
project(FUSEHello C)

set(CMAKE_C_STANDARD 11)
set (CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH}
        "${CMAKE_SOURCE_DIR}/CMakeModules/")
set(CMAKE_C_FLAGS "-D_FILE_OFFSET_BITS=64")

find_package(FUSE REQUIRED)
include_directories(${FUSE_INCLUDE_DIR})

add_executable(FUSEHello main.c)

target_link_libraries(FUSEHello ${FUSE_LIBRARIES})
