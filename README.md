# keeprun
Keep console program running.

## Install

    go get -v github.com/codeskyblue/keeprun

## Usage
    Usage of ./keeprun:
      -delay duration
            Delay between each restart (default 5s)
      -killon string
            Kill program when text appear

    $ keeprun sleep 50
    # restart sleep again when sleep down

When keeprun get SIGTERM, SIGHUP signal, process `sleep` will also be killed.

## LICENSE
[MIT](LICENSE)
