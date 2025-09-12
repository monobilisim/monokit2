#!/usr/bin/env bash

case $1 in
    "build")
        case $2 in
            "osHealth")
                if [ -f plugins/bin/osHealth ]; then
                    rm plugins/bin/osHealth
                fi
                cd plugins/osHealth
                go build -ldflags "-X 'main.version=devel'" -tags osHealth -o ../bin/
                cd ../..
                ;;
            *)
                if [ -f ./bin/monokit2 ]; then
                    rm ./bin/monokit2
                fi
                # if no plugin is selected, build the main application
                go build -ldflags "-X 'main.version=devel'" -o ./bin/monokit2 ./main.go
        esac
        ;;
    "run")
        case $2 in
            "osHealth")
                if [ -f plugins/bin/osHealth ]; then
                    rm plugins/bin/osHealth
                fi
                cd plugins/osHealth
                go build -ldflags "-X 'main.version=devel'" -tags osHealth -o ../bin/
                cd ../..
                ./plugins/bin/osHealth "${@:3}"
                ;;
            *)
                if [ -f ./bin/monokit2 ]; then
                    rm ./bin/monokit2
                fi
                # if no plugin is selected, build and run the main application
                go build -ldflags "-X 'main.version=devel'" -o ./bin/monokit2 ./main.go
                ./bin/monokit2 "${@:3}"
        esac
        ;;
    "send")
        if [ -z "$2" ]; then
            echo "Usage: $0 send user@host [plugin]"
            exit 1
        fi

        ssh "$2" "mkdir -p /var/lib/monokit2/plugins"

        case $3 in
            "osHealth")
                scp ./plugins/bin/$3 $2:/var/lib/monokit2/plugins
                ;;
            *)
                scp ./bin/monokit2 $2:/usr/local/bin
        esac
        ;;
    *)
        echo "Usage: $0 {build|run|send} [plugin]"
        echo "If no plugin is specified, the main application will be built or run."
        echo "Everything after the {build|run} will be passed as arguments to the application."
        echo "send user@host [plugin] will copy the binary to the remote host plugins."
        echo "If no plugin is specified, the main application will be copied to host."
        echo "Available plugins: osHealth"
        ;;
esac
