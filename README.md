# zap-graylog-writer

Writes log message from uber-go/zap to GELF input (as provided by graylog2, ELK stack, etc.). 
Code for feeding data into gelf sink has been taken from [github.com/robertkowalski/graylog-golang](https://github.com/robertkowalski/graylog-golang).

Please ensure that the [required](http://docs.graylog.org/en/2.1/pages/gelf.html) gelf fields are provided. 

Example for setting up zap with gelf data sink:

    var logger zap.Logger
    var gelfForwarder zap.WriteSyncer
    func init() {
        cfg := gelf.Config{GraylogHostname: "192.168.1.6", GraylogPort: 12201, MaxChunkSize: 8154}
        gelfForwarder = gelf.New(cfg)
        lvf := func() zap.LevelFormatter {
            return zap.LevelFormatter(func(l zap.Level) zap.Field {
                return zap.Int(gelf.LevelTag, gelf.ZapLevelToGelfLevel(int32(l)))
            })
        }
        logger = zap.New(
            zap.NewJSONEncoder(zap.NoTime(), zap.MessageFormatter(zap.MessageKey("short_message")), lvf()),
            zap.DebugLevel,
            zap.Fields(
                zap.String(gelf.VersionTag, "1.1"),
                zap.String(gelf.HostTag, "publisher"),
                zap.Int("_pid", os.Getpid()),
                zap.String("_exe", path.Base(os.Args[0]))),
            zap.Output(gelfForwarder),
        )
    }

## Notes

* currently compression not supported!
* no safety checks for invalid fields!
* use at your own risk :)
