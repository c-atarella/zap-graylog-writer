# zap-graylog-writer

Writes log messages from go.uber.org/zap to GELF input (as provided by [graylog2](http://docs.graylog.org/en/2.1/pages/sending_data.html), ELK stack, etc.).
Code for feeding data into gelf sink has been taken from [github.com/robertkowalski/graylog-golang](https://github.com/robertkowalski/graylog-golang).

A complete example for setting up zap with gelf data sink:

    package main

    import (
        "os"
        "path"

        "github.com/c-atarella/zap-graylog-writer"
        "go.uber.org/zap"
    )

    const (
        appVersion = "0.0.1"
    )

    var baseLogger *zap.Logger

    func init() {
        core, fieldsOption := gelf.NewGelfCore("192.168.1.6", "my_fancy_app",
            zap.Int("_pid", os.Getpid()),
            zap.String("_exe", path.Base(os.Args[0])),
            zap.String("_appversion", appVersion))
        baseLogger = zap.New(core, fieldsOption)
        defer baseLogger.Sync()

    }

    func main() {
        for i := 0; i < 10; i++ {
            baseLogger.Info("huh jo", zap.Int("MessageId", i))
        }
    }

## Notes

* currently compression is not supported!
* no safety checks for invalid fields!
* use at your own risk :)
* tested with graylog-2.2.3