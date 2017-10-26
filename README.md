# XLOG

Simple usage: 

----
# **Console**: 

 Write log at console with colors
 

#### Config Options:

|  Variables     | Types        | Description                                | Options |
|----------------|--------------|--------------------------------------------|---------
|	Level        | LEVEL        | Define level of log (Default: **0**)       | TRACE, INFO, WARN, ERROR, FATAL
|	BufferSize   | int64        | Buffer size (Default: **0**)               | In64 size
|	Writer       | io.Writer    | Where will write log (Default: **Stdout**) | Any io.writer implementation
|   ValuesDefault| interface    | Define default values for output           | array of interface{}| 


#### Example: 

```go
err := xlog.New(xlog.CONSOLE, xlog.ConsoleConfig{})
if err != nil {
    fmt.Printf("Fail to create new logger: %v\n", err)
    os.Exit(1)
}

now := time.Now().Format("02 Jan 06 15:04")

xlog.Info("type", "Info", "Transponder", now, "version", xlog.Version())
xlog.Trace("type", "Trace", "Transponder", now, "version", xlog.Version())
```

####  Output: 

```bash
[ INFO] service=XPTO type=Info Transponder="25 Oct 17 18:11" version=0.0.1
[TRACE] service=XPTO type=Trace Transponder="25 Oct 17 18:11" version=0.0.1
```


----
# **Json Format**:

#### Config Options:

|  Variables     | Types        | Description                                | Options |
|----------------|--------------|--------------------------------------------|---------
|	Level        | LEVEL        | Define level of log (Default: **0**)       | TRACE, INFO, WARN, ERROR, FATAL
|	BufferSize   | int64        | Buffer size (Default: **0**)               | In64 size
|	Writer       | io.Writer    | Where will write log (Default: **Stdout**) | Any io.writer implementation
|   ValuesDefault| interface    | Define default values for output           | array of interface{}|

####  Example:

```go
err := xlog.New(xlog.JSONFormat, xlog.JsonFormatConfig{
    Writer: os.Stdout,
    Level: xlog.TRACE,
    ValuesDefault: []interface{}{"service", "XPTO"},
})
if err != nil {
    fmt.Printf("Fail to create new logger: %v\n", err)
    os.Exit(1)
}

now := time.Now().Format("02 Jan 06 15:04")

xlog.Info("type", "Info", "Transponder", now, "version", xlog.Version())
xlog.Trace("type", "Trace", "Transponder", now, "version", xlog.Version()
```

####  Output: 

```json
{"service": "XPTO", "Transponder":"25 Oct 17 18:13","type":"Info","version":"0.0.1"}
{"service": "XPTO", "Transponder":"25 Oct 17 18:13","type":"Trace","version":"0.0.1"}
```

---

