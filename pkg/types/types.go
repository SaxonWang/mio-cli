package types

const (
	HIADMIN_DOMAIN string = "hiadmin-demo.app.vpclub.io"
	TOKEN_API          string = "http://"+HIADMIN_DOMAIN+"/login"
	PIPELINE_START_API string = "http://"+HIADMIN_DOMAIN+"/pipelineConfig/starter"
	BUILDLOG_API string = "ws://"+HIADMIN_DOMAIN+"/websocket/buildLogs"

	TOKEN_DIR  string = "token"
	TOKEN_FILE string = "hidevopsio"
)

var Message = make(chan string)

type PipelineStart struct {
	Name       string
	Namespace  string
	SourceCode string
}
