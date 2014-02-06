function check_exit {
    if [ $? -ne 0 ]; then
        echo "ERROR: Build failed ...."
        exit
    fi
    echo
}

commands=( "commands"
           "connect http://localhost:9000 default beer-sample"
           "help"
           "help connect"
           'set "testkey" 0 "test value"'
           'get "testkey"'
           "list nodes"
           "list pools"
           "list buckets" )

# Build cbsh
go build ./...
check_exit

for index in ${!commands[*]}; do
    cmd="${commands[$index]}"
    echo "---------- $cmd -------------"
    ./cbsh -url http://localhost:9000 -cmd "$cmd"
    check_exit
done
