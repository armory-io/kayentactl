package kayenta

// StdLogger defines a really basic log interface that
// users can implement to provide a custom log handler
type StdLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

type NoopStdLogger struct{}

func (n NoopStdLogger) Debugf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Infof(format string, args ...interface{}) {

}

func (n NoopStdLogger) Printf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Warnf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Warningf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Errorf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Fatalf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Panicf(format string, args ...interface{}) {

}

func (n NoopStdLogger) Debug(args ...interface{}) {

}

func (n NoopStdLogger) Info(args ...interface{}) {

}

func (n NoopStdLogger) Print(args ...interface{}) {

}

func (n NoopStdLogger) Warn(args ...interface{}) {

}

func (n NoopStdLogger) Warning(args ...interface{}) {

}

func (n NoopStdLogger) Error(args ...interface{}) {

}

func (n NoopStdLogger) Fatal(args ...interface{}) {

}

func (n NoopStdLogger) Panic(args ...interface{}) {

}

func (n NoopStdLogger) Debugln(args ...interface{}) {

}

func (n NoopStdLogger) Infoln(args ...interface{}) {

}

func (n NoopStdLogger) Println(args ...interface{}) {

}

func (n NoopStdLogger) Warnln(args ...interface{}) {

}

func (n NoopStdLogger) Warningln(args ...interface{}) {

}

func (n NoopStdLogger) Errorln(args ...interface{}) {

}

func (n NoopStdLogger) Fatalln(args ...interface{}) {

}

func (n NoopStdLogger) Panicln(args ...interface{}) {

}
