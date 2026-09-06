cmake_minimum_required(VERSION 3.10)
project(nexus-core LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

# Source files
set(RATELIMIT_SRC
    ratelimit/token_bucket.cpp
    ratelimit/sliding_window.cpp
    ratelimit/max_concurrency.cpp
)

set(PROC_READER_SRC
    proc_reader/cpu.cpp
    proc_reader/mem.cpp
    proc_reader/disk.cpp
    proc_reader/net.cpp
)

set(BRIDGE_SRC
    ${CMAKE_SOURCE_DIR}/../pkg/cbridge/bridge.cpp
)

# Static library
add_library(nexus_core STATIC
    ${RATELIMIT_SRC}
    ${PROC_READER_SRC}
    ${BRIDGE_SRC}
)

target_include_directories(nexus_core PUBLIC
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_SOURCE_DIR}/ratelimit
    ${CMAKE_SOURCE_DIR}/proc_reader
    ${CMAKE_SOURCE_DIR}/../pkg/cbridge
)

# Install
install(TARGETS nexus_core DESTINATION lib)
install(DIRECTORY include/ DESTINATION include)

# Tests (optional)
option(BUILD_TESTS "Build unit tests" OFF)
if(BUILD_TESTS)
    enable_testing()
    find_package(GTest REQUIRED)

    add_executable(test_token_bucket tests/test_token_bucket.cpp)
    target_link_libraries(test_token_bucket nexus_core GTest::GTest GTest::Main)
    add_test(NAME TokenBucket COMMAND test_token_bucket)

    add_executable(test_sliding_window tests/test_sliding_window.cpp)
    target_link_libraries(test_sliding_window nexus_core GTest::GTest GTest::Main)
    add_test(NAME SlidingWindow COMMAND test_sliding_window)

    add_executable(test_max_concurrency tests/test_max_concurrency.cpp)
    target_link_libraries(test_max_concurrency nexus_core GTest::GTest GTest::Main)
    add_test(NAME MaxConcurrency COMMAND test_max_concurrency)

    add_executable(test_proc_reader tests/test_proc_reader.cpp)
    target_link_libraries(test_proc_reader nexus_core GTest::GTest GTest::Main)
    add_test(NAME ProcReader COMMAND test_proc_reader)
endif()
