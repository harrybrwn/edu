module github.com/harrybrwn/edu

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/gen2brain/beeep v0.0.0-20200526185328-e9c15c258e28
	github.com/harrybrwn/config v0.1.2
	github.com/harrybrwn/errs v0.0.2-0.20200523142445-e4279967174e
	github.com/harrybrwn/go-canvas v0.0.2-0.20200821044925-04b08bcecf29
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/mitchellh/mapstructure v1.3.3
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/harrybrwn/go-canvas => ../../pkg/go-canvas
