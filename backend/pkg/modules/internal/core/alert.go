package core

const AlertAnnotation = "ALERT"

type AlertSeverity int

const (
	AlertInfo AlertSeverity = iota
	AlertWarn
)

var (
	AlertInfoAnn = Annotation{Name: AlertAnnotation, Annotation: []byte(AlertInfo.String())}
	AlertWarnAnn = Annotation{Name: AlertAnnotation, Annotation: []byte(AlertWarn.String())}
)

func (es AlertSeverity) String() string {
	switch es {
	case AlertInfo:
		return "ALERT_INFO"
	case AlertWarn:
		return "ALERT_WARN"
	}
	panic("undefined alert severity")
}
