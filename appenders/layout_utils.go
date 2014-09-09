package appenders

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dspasibenko/log4g"
)

// layout pieces types
const (
	lpText = iota
	lpLoggerName
	lpDate
	lpLogLevel
	lpMessage
)

// parse states
const (
	psText = iota
	psPiece
	psDateStart
	psDate
)

type layoutPiece struct {
	value     string
	pieceType int
}

type LayoutTemplate []layoutPiece

var logLevelNames []string = log4g.LevelNames()

/**
 * Parses log message layout template with the following placeholders:
 * %c - logger name
 * %d{date/time format} - date/time. The date/time format should be specified
 *				in time.Format() form like "Mon, 02 Jan 2006 15:04:05 -0700"
 * %p - priority name
 * %m - the log message
 * %% - '%'
 */
func ParseLayout(layout string) (LayoutTemplate, error) {
	layoutTemplate := make(LayoutTemplate, 0, 10)
	state := psText
	startIdx := 0
	for i, rune := range layout {
		switch state {
		case psText:
			if rune == '%' {
				layoutTemplate = addPiece(layout[startIdx:i], lpText, layoutTemplate)
				state = psPiece
			}
		case psPiece:
			state = psText
			startIdx = i + 1
			switch rune {
			case 'c':
				layoutTemplate = addPiece("c", lpLoggerName, layoutTemplate)
			case 'd':
				state = psDateStart
			case 'p':
				layoutTemplate = addPiece("p", lpLogLevel, layoutTemplate)
			case 'm':
				layoutTemplate = addPiece("m", lpMessage, layoutTemplate)
			case '%':
				startIdx = i
			default:
				return nil, errors.New("Unknown layout identifier " + string(rune))
			}
		case psDateStart:
			if rune != '{' {
				return nil, errors.New("%d should follow by date format in braces like this: %d{...}, but found " + string(rune))
			}
			startIdx = i + 1
			state = psDate
		case psDate:
			if rune == '}' {
				layoutTemplate = addPiece(layout[startIdx:i], lpDate, layoutTemplate)
				state = psText
				startIdx = i + 1
			}
		}
	}

	if state != psText {
		return nil, errors.New("Unexpected end of layout, cannot parse it properly")
	}

	return addPiece(layout[startIdx:len(layout)], lpText, layoutTemplate), nil
}

func ToLogMessage(logEvent *log4g.LogEvent, template LayoutTemplate) string {
	buf := bytes.NewBuffer(make([]byte, 0, 64))

	for _, piece := range template {
		switch piece.pieceType {
		case lpText:
			buf.WriteString(piece.value)
		case lpLoggerName:
			buf.WriteString(logEvent.LoggerName)
		case lpDate:
			buf.WriteString(logEvent.Timestamp.Format(piece.value))
		case lpLogLevel:
			buf.WriteString(logLevelNames[logEvent.Level])
		case lpMessage:
			buf.WriteString(fmt.Sprint(logEvent.Payload))
		}
	}
	return buf.String()
}

func addPiece(str string, pieceType int, template LayoutTemplate) LayoutTemplate {
	if len(str) == 0 {
		return template
	}
	return append(template, layoutPiece{str, pieceType})
}